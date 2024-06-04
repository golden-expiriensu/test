package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	const numberOfQueue = 2
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var wg sync.WaitGroup

	// ready to receive messages from clients. use RabbitMQ client to publish messages and see the result in the terminal
	for i := 0; i < numberOfQueue; i++ {
		wg.Add(i)
		if i > len(alphabet) {
			log.Panic("Number of queue cannot exceed 26")
		}

		go func(i int) {
			inputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), "input-"+string(alphabet[i]))
			if err != nil {
				log.WithError(err).Panic("Cannot create input queue")
			}

			outputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), "output-"+string(alphabet[i]))
			if err != nil {
				log.WithError(err).Panic("Cannot create output queue")
			}

			if err = processor.New(inputQueue, outputQueue, database.D{}).Run(ctx); err != nil {
				log.WithError(err).Panic("Error running processor " + string(alphabet[i]))
			}
		}(i)
	}

	log.Info("Application is ready to run")

	wg.Wait()
}
