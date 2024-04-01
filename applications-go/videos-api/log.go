package main

import (
	"log"
	"os"
)

func NewLogger() {
	var infoFilename string = "/tmp/videos-api-info.log"
	infoFile, err := os.Open(infoFilename)
	if err != nil {
		log.Fatal("Couldn't open", infoFile)
	}
	infoLogger = log.New(infoFile, "INFO:", log.Lshortfile)

	var errFilename string = "/tmp/videos-api-err.log"
	errFile, err := os.Open(errFilename)
	if err != nil {
		log.Fatal("Couldn't open", errFile)
	}
	errLogger = log.New(errFile, "ERROR:", log.Lshortfile)
}
