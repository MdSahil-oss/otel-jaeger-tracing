package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setHttpRequest() {
	infoLogger.Println("Setting for http request")
	infoLogger.Println("Configuring Jaeger")
	cfg := newJaegerConfig()

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		errLogger.Panicln("ERROR: cannot init Jaeger: ", err)
	}
	infoLogger.Println("Configured Jaeger")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	router := httprouter.New()

	router.GET("/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		infoLogger.Println("Sent GET request at '/id'")
		spanCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)

		span := tracer.StartSpan("videos-api: GET /id", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		if flaky == "true" {
			if rand.Intn(90) < 30 {
				errLogger.Panicln("flaky error occurred ")
			}
		}

		ctx := opentracing.ContextWithSpan(ctx, span)
		videos := getVideo(w, r, p, ctx)

		if strings.Contains(videos[0].Id, "jM36M39MA3I") && delay == "true" {
			time.Sleep(6 * time.Second)
		}

		jsonData, err := json.Marshal(videos[0])
		if err != nil {
			errLogger.Panicln(err)
		}

		cors(w)
		fmt.Fprintf(w, "%s", string(jsonData))
	})

	errLogger.Fatal(http.ListenAndServe(":10010", router))
}

func getVideo(writer http.ResponseWriter, request *http.Request, p httprouter.Params, ctx context.Context) (videos []video) {
	infoLogger.Println("Fetching data from database")
	span, _ := opentracing.StartSpanFromContext(ctx, "videos-api: mongo-get")
	defer span.Finish()
	id := p.ByName("id")

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongo_user+":"+mongo_password+"@"+mongo_host+":"+mongo_port))
	if err != nil {
		errLogger.Panicln(err)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			errLogger.Panicln(err)
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
		errLogger.Panicln(err)
	}

	if err = cursor.All(ctx, &videos); err != nil {
		errLogger.Panicln(err)
	} else {
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(request.Header),
		)
	}
	infoLogger.Println("Fetched data from database")
	return videos
}
