package main

import (
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/acknode/ackstream/cmd"
	"github.com/acknode/ackstream/pkg/configs"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f:", r)
			log.Println("Stack trace:", string(debug.Stack()))
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
		log.Println("stopping...")
		time.Sleep(5 * time.Second)
	}

	os.Exit(code)
}
