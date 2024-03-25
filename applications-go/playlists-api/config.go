package main

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
