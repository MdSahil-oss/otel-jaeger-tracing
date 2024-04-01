package main

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getPlaylists(ctx context.Context) (playlists []playlist) {
	infoLogger.Println("Fetching data from database")
	span, _ := opentracing.StartSpanFromContext(ctx, "playlists-api: mongo-get")
	defer span.Finish()

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
	cursor, err := coll.Find(ctx, bson.D{}, options.Find())
	if err != nil {
		errLogger.Panicln(err)
	}

	if err = cursor.All(ctx, &playlists); err != nil {
		errLogger.Panicln(err)
	}
	infoLogger.Println("Fetched data from database")
	return playlists
}
