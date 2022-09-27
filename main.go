package main

import (
	"log"
	"os"
	"time"

	"github.com/acknode/ackstream/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f:", r)
			exit(2)
		}
	}()

	command := cmd.New()
	if err := command.Execute(); err != nil {
		log.Print(err)
		exit(1)
	}
}

func exit(code int) {
	prod := os.Getenv("ACKSTREAM_ENV") != "dev"
	if prod {
		time.Sleep(5 * time.Second)
	}

	os.Exit(code)
}
