// Package main implements the example.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/refresh/refresh"
)

func main() {

	amqpURL := "amqp://guest:guest@rabbitmq:5672/"
	me := filepath.Base(os.Args[0])
	debug := true

	// "#" means receive notification for all applications
	refresher := refresh.New(amqpURL, me, []string{"#"}, debug, nil)

	for app := range refresher.C {
		log.Printf("refresh: received notification for application='%s'", app)
	}
}
