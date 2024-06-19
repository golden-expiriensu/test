package main

import (
	"context"
	"os"

	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	inputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), "input-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create input queue")
	}

	outputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), "output-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create output queue")
	}

	log.Info("Application is ready to run")

	err = processor.New(inputQueue, outputQueue, database.D{}).Run(ctx)
	if err != nil {
		log.WithError(err).Error("Running processor")
	}
}
