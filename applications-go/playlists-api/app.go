package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

const serviceName = "playlists-api"

var environment = os.Getenv("ENVIRONMENT")
var jaeger_host_port = os.Getenv("JAEGER_HOST_PORT")

var ctx = context.Background()
var mongo_host = os.Getenv("MONGO_HOST")
var mongo_port = os.Getenv("MONGO_PORT")
var mongo_user = os.Getenv("MONGO_USER")
var mongo_password = os.Getenv("MONGO_PASSWORD")
var mongoUri = "mongodb://" + mongo_user + ":" + mongo_password + "@" + mongo_host + ":" + mongo_port

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

	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		spanCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)

		span := tracer.StartSpan("playlists-api: GET /", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		cors(w)

		ctx := opentracing.ContextWithSpan(ctx, span)
		playlists := getPlaylists(ctx)
		//get videos for each playlist from videos api
		for pi := range playlists {

			vs := []video{}
			for vi := range playlists[pi].Videos {

				span, _ := opentracing.StartSpanFromContext(ctx, "playlists-api: videos-api GET /id")

				v := video{}

				req, err := http.NewRequest("GET", "http://videos-api:10010/"+playlists[pi].Videos[vi].Id, nil)
				if err != nil {
					panic(err)
				}

				span.Tracer().Inject(
					span.Context(),
					opentracing.HTTPHeaders,
					opentracing.HTTPHeadersCarrier(req.Header),
				)

				videoResp, err := http.DefaultClient.Do(req)
				span.Finish()

				if err != nil {
					fmt.Println(err)
					span.SetTag("error", true)
					break
				}

				defer videoResp.Body.Close()
				video, err := ioutil.ReadAll(videoResp.Body)

				if err != nil {
					panic(err)
				}

				err = json.Unmarshal(video, &v)

				if err != nil {
					panic(err)
				}

				vs = append(vs, v)

			}

			playlists[pi].Videos = vs
		}

		playlistsBytes, err := json.Marshal(playlists)
		if err != nil {
			panic(err)
		}

		reader := bytes.NewReader(playlistsBytes)
		if b, err := ioutil.ReadAll(reader); err == nil {
			fmt.Fprintf(w, "%s", string(b))
		}

	})

	fmt.Println("Running...")
	log.Fatal(http.ListenAndServe(":10010", router))
}

func getPlaylists(ctx context.Context) (playlists []playlist) {

	span, _ := opentracing.StartSpanFromContext(ctx, "playlists-api: mongo-get")
	defer span.Finish()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := mongoClient.Database("test").Collection("data")
	cursor, err := coll.Find(ctx, bson.D{}, options.Find())
	if err != nil {
		panic(err)
	}

	if err = cursor.All(ctx, &playlists); err != nil {
		panic(err)
	}

	return playlists
}

type playlist struct {
	Id     string  `bson:"id" json:"id"`
	Name   string  `bson:"name" json:"name"`
	Videos []video `bson:"videos" json:"videos"`
}

type video struct {
	Id          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Imageurl    string `json:"imageurl" bson:"imageurl"`
	Url         string `json:"url" bson:"url"`
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
