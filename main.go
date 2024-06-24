package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	inputQueues := []string{"input-A", "input-B"}
	outputQueues := []string{"output-A", "output-B"}

	for i, inputQueueName := range inputQueues {
		outputQueueName := outputQueues[i]

		inputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), inputQueueName)
		if err != nil {
			log.WithError(err).Panicf("Cannot create input queue %s", inputQueueName)
		}

		outputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), outputQueueName)
		if err != nil {
			log.WithError(err).Panicf("Cannot create output queue %s", outputQueueName)
		}

		log.Infof("Application is ready to run for %s -> %s", inputQueueName, outputQueueName)

		go processor.New(inputQueue, outputQueue, database.D{}).Run(ctx)
	}

	// Block forever
	select {}
}
