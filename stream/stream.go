package stream

import (
	"context"
	"os"

	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"

	log "github.com/sirupsen/logrus"
)

var InputStreams = make(map[string]processor.Queue)
var OutputStreams = make(map[string]<-chan queue.Delivery)

var available = make(map[string]bool)

func New(name string, ctx context.Context) (success bool, inputName string, outputName string) {
	if available[name] {
		log.WithField("name", name).Error("Stream already exists")
		return
	}

	inputName = "input-" + name
	outputName = "output-" + name

	inputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), inputName)
	if err != nil {
		log.WithError(err).Panic("Cannot create input queue")
	}

	outputQueue, err := queue.New(os.Getenv("RABBITMQ_URL"), outputName)
	if err != nil {
		log.WithError(err).Panic("Cannot create output queue")
	}

	go processor.New(inputQueue, outputQueue, database.D{}).Run(ctx)

	success = true
	available[name] = success
	InputStreams[inputName] = inputQueue
	OutputStreams[outputName], err = outputQueue.Consume(ctx)

	if err != nil {
		log.WithError(err).Panic("Cannot consume from output queue for ", outputName)
	}

	return
}
