package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func setHttpRequest() {
	cfg := newJaegerConfig()

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
		fmt.Println("Playlists:", playlists)
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
	log.Fatal(http.ListenAndServe(":10010", router))
}

func cors(writer http.ResponseWriter) {
	if environment == "DEBUG" {
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-MY-API-Version")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
	}
}
