package main

import (
	"context"
	"fmt"
	"os"
)

const serviceName = "playlists-api"

var environment = os.Getenv("ENVIRONMENT")
var jaeger_host_port = os.Getenv("JAEGER_HOST_PORT")

var ctx = context.Background()
var mongo_host = os.Getenv("MONGO_HOST")
var mongo_port = os.Getenv("MONGO_PORT")
var mongo_user = os.Getenv("MONGO_USER")
var mongo_password = os.Getenv("MONGO_PASSWORD")
var vault_addr = os.Getenv("VAULT_ADDR")
var jwt_path = os.Getenv("JWT_PATH")
var mongo_db = "test"
var mongo_collection = "playlists"

func main() {
	fetchSecretsFromVault()
	setHttpRequest()
	fmt.Println("Running...")
}
