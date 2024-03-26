package main

import "net/http"

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
