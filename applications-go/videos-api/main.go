package main

import (
	"context"
	"os"

	"log"
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
var infoLogger, errLogger *log.Logger

var ctx = context.Background()

func main() {
	NewLogger()
	fetchSecretsFromVault()
	setHttpRequest()
	infoLogger.Println("Running...")
}
