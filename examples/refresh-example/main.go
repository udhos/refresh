// Package main implements the example.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/refresh/refresh"
)

func main() {

	amqpURLEnv := "AMQP_URL"
	amqpURL := "amqp://guest:guest@rabbitmq:5672/"
	amqpURLValue := os.Getenv(amqpURLEnv)
	if amqpURLValue != "" {
		amqpURL = amqpURLValue
	}

	log.Printf("%s='%s' using amqpURL='%s'", amqpURLEnv, amqpURLValue, amqpURL)

	me := filepath.Base(os.Args[0])
	debug := true

	options := refresh.Options{
		AmqpURL:      amqpURL,
		ConsumerTag:  me,
		Applications: []string{"#"},
		Debug:        debug,
	}

	// "#" means receive notification for all applications
	refresher := refresh.New(options)

	for app := range refresher.C {
		log.Printf("refresh: received notification for application='%s'", app)
	}

	refresher.Close()
}
