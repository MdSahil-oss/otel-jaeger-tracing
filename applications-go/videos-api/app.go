package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//TODO : https://opentracing.io/guides/golang/quick-start/
// docker run --rm -it -p 6831:6831/udp -p 16686:16686 -p 14269:14269  --net tracing --name jaeger jaegertracing/all-in-one:latest

const serviceName = "videos-api"

var environment = os.Getenv("ENVIRONMENT")
var jaeger_host_port = os.Getenv("JAEGER_HOST_PORT")
var flaky = os.Getenv("FLAKY")
var delay = os.Getenv("DELAY")
var vault_addr = os.Getenv("VAULT_ADDR")
var jwt_path = os.Getenv("JWT_PATH")
var mongo_host = os.Getenv("MONGO_HOST")
var mongo_port = os.Getenv("MONGO_PORT")
var mongo_user = os.Getenv("MONGO_USER")
var mongo_password = os.Getenv("MONGO_PASSWORD")
var mongo_db = "test"
var mongo_collection = "videos"

var ctx = context.Background()

func main() {
	cfg := &config.Configuration{
		ServiceName: serviceName,

		// "const" sampler is a binary sampling strategy: 0=never sample, 1=always sample.
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		// Log the emitted spans to stdout.
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaeger_host_port,
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	if len(mongo_user) == 0 && len(mongo_password) == 0 {
		config := vault.DefaultConfig() // modify for more granular configuration
		config.Address = vault_addr
		client, err := vault.NewClient(config)
		if err != nil {
			panic(fmt.Errorf("unable to initialize Vault client: %w", err))
		}

		k8sAuth, err := auth.NewKubernetesAuth(
			"mongodb",
			auth.WithServiceAccountTokenPath(jwt_path),
		)
		if err != nil {
			panic(fmt.Errorf("unable to initialize Kubernetes auth method: %w", err))
		}

		authInfo, err := client.Auth().Login(ctx, k8sAuth)
		if err != nil {
			panic(fmt.Errorf("unable to log in with Kubernetes auth: %w", err))
		}
		if authInfo == nil {
			panic(fmt.Errorf("no auth info was returned after login"))
		}

		// get secret from Vault, from the default mount path for KV v2 in dev mode, "secret"
		secret, err := client.KVv2("secret").Get(context.Background(), "mongodb/config")
		if err != nil {
			panic(fmt.Errorf("unable to read secret: %w", err))
		}

		// data map can contain more than one key-value pair,
		// in this case we're just grabbing one of them
		username, ok := secret.Data["username"].(string)
		if !ok {
			panic(fmt.Sprintf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"]))
		}
		password, ok := secret.Data["password"].(string)
		if !ok {
			panic(fmt.Sprintf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"]))
		}

		mongo_user = strings.Trim(username, " ")
		mongo_password = strings.Trim(password, " ")
	}

	router := httprouter.New()

	router.GET("/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		spanCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)

		span := tracer.StartSpan("videos-api: GET /id", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		if flaky == "true" {
			if rand.Intn(90) < 30 {
				panic("flaky error occurred ")
			}
		}

		ctx := opentracing.ContextWithSpan(ctx, span)
		videos := getVideo(w, r, p, ctx)

		if strings.Contains(videos[0].Id, "jM36M39MA3I") && delay == "true" {
			time.Sleep(6 * time.Second)
		}

		jsonData, err := json.Marshal(videos[0])
		if err != nil {
			panic(err)
		}

		cors(w)
		fmt.Fprintf(w, "%s", string(jsonData))
	})

	fmt.Println("Running...")
	log.Fatal(http.ListenAndServe(":10010", router))
}

func getVideo(writer http.ResponseWriter, request *http.Request, p httprouter.Params, ctx context.Context) (videos []video) {

	span, _ := opentracing.StartSpanFromContext(ctx, "videos-api: mongo-get")
	defer span.Finish()
	id := p.ByName("id")

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongo_user+":"+mongo_password+"@"+mongo_host+":"+mongo_port))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := mongoClient.Database(mongo_db).Collection(mongo_collection)
	cursor, err := coll.Find(ctx, bson.D{{"id", id}}, options.Find())
	if err == mongo.ErrNoDocuments {
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(request.Header),
		)
		return
	} else if err != nil {
		panic(err)
	}

	if err = cursor.All(ctx, &videos); err != nil {
		panic(err)
	} else {
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(request.Header),
		)
	}

	return videos
}

type stop struct {
	error
}

func cors(writer http.ResponseWriter) {
	if environment == "DEBUG" {
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-MY-API-Version")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
	}
}

type video struct {
	Id          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Imageurl    string `json:"imageurl" bson:"imageurl"`
	Url         string `json:"url" bson:"url"`
}
