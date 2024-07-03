package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"os"

	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	database := database.D{}

	processingGroup, processingGroupContext := errgroup.WithContext(ctx)
	for _, queueNo := range [...]string{"A", "B"} {
		inputQueueName := "input-" + queueNo
		inputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), inputQueueName)
		if err != nil {
			log.WithError(err).Panic("Cannot create input queue")
		}

		outputQueueName := "output-" + queueNo
		outputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), outputQueueName)
		if err != nil {
			log.WithError(err).Panic("Cannot create output queue")
		}

		log.Infof("Application %s is ready to run\n", queueNo)

		processingGroup.Go(
			func() error {
				return processor.New(inputQueue, outputQueue, database).Run(processingGroupContext)
			},
		)
	}

	if err := processingGroup.Wait(); err != nil {
		log.WithError(err).Panicf("failed to wait for processing group: %v", err)
	}

	log.Info("Graceful shutdown.")
}
