package main

import (
	"github.com/acknode/ackstream/utils"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/acknode/ackstream/cmd"
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
	if isDebug := utils.IsDebug("ACKSTREAM_ENV"); !isDebug {
		log.Println("stopping...")
		time.Sleep(5 * time.Second)
	}

	os.Exit(code)
}
