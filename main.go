package main

import (
	"log"
	"os"
)

func main() {
	defer logging().Close()

	app := NewApp()
	err := app.Init()
	if err != nil {
		panic(err)
	}
}

func logging() *os.File {
	LOG_FILE := "./log.txt"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(logFile)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return logFile
}
