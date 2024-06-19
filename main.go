package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
)

type stream struct {
	inputQueueName  string
	outputQueueName string
	database        processor.Database
}

var streams = map[string]stream{
	"A": {
		inputQueueName:  "input-A",
		outputQueueName: "output-A",
		database:        database.D{},
	},
	"B": {
		inputQueueName:  "input-B",
		outputQueueName: "output-B",
		database:        database.D{},
	},
}

func main() {
	ctx := context.Background()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Panic("RABBITMQ_URL environment variable not set")
	}

	log.Info("Application is ready to run")

	err := startStreamProcessors(ctx, rabbitURL)
	if err != nil {
		log.WithField("reason", err).Info("Exiting")
	}
}

func startStreamProcessors(ctx context.Context, connectionURL string) error {
	wg := sync.WaitGroup{}
	for strName, str := range streams {
		wg.Add(1)
		go func(name string, s stream) {
			defer wg.Done()

			inputQueue, err := queue.New(connectionURL, s.inputQueueName)
			if err != nil {
				log.WithError(err).Errorf("Create input queue for stream %s", name)
				return
			}
			outputQueue, err := queue.New(connectionURL, s.outputQueueName)
			if err != nil {
				log.WithError(err).Errorf("Create output queue for stream %s", name)
				return
			}
			err = processor.New(inputQueue, outputQueue, s.database).Run(ctx)
			if err != nil {
				log.WithError(err).Errorf("Running processor for stream %s", name)
				return
			}
		}(strName, str)
	}

	wg.Wait()
	return fmt.Errorf("all processors failed")
}
