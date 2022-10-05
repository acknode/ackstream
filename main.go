package main

import (
	"log"
	"os"
	"time"

	"github.com/acknode/ackstream/cmd"
	"github.com/acknode/ackstream/pkg/configs"
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
	if debug := configs.IsDebug("ACKSTREAM_ENV"); !debug {
		time.Sleep(5 * time.Second)
	}

	os.Exit(code)
}
