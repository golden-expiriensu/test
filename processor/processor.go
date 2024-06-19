package processor

import (
	"context"

	"arkis_test/queue"

	log "github.com/sirupsen/logrus"
)

type Queue interface {
	Consume(ctx context.Context) (<-chan queue.Delivery, error)
	Publish(ctx context.Context, msg []byte) error
}

type Database interface {
	Get([]byte) (string, error)
}

type Processor struct {
	input    Queue
	output   Queue
	database Database
}

func New(input, output Queue, db Database) Processor {
	return Processor{input, output, db}
}

func (p Processor) Run(ctx context.Context) error {
	deliveries, err := p.input.Consume(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery := <-deliveries:
			if err := p.process(ctx, delivery); err != nil {
				return err
			}
		}
	}
}

func (p Processor) process(ctx context.Context, delivery queue.Delivery) error {
	log.WithField("delivery", string(delivery.Body)).Info("Processing the delivery")

	data, err := p.database.Get(delivery.Body)
	if err != nil {
		return err
	}

	log.WithField("result", data).Info("Processed the delivery")

	return p.output.Publish(ctx, []byte(data))
}
