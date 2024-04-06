package main

import (
	"log"
	"os"
)

func NewLogger() {
	var infoFilename string = "/app/videos-api-info.log"
	infoFile, err := os.Create(infoFilename)
	if err != nil {
		log.Fatal("Couldn't create", err)
	}
	infoLogger = log.New(infoFile, "INFO:", log.Lshortfile)

	var errFilename string = "/app/videos-api-err.log"
	errFile, err := os.Create(errFilename)
	if err != nil {
		log.Fatal("Couldn't create", err)
	}

	errLogger = log.New(errFile, "ERROR:", log.Lshortfile)
}
