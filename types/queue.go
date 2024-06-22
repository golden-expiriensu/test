package types

import (
	"arkis_test/queue"
	"context"
)

type Queue interface {
	Consume(ctx context.Context) (<-chan queue.Delivery, error)
	Publish(ctx context.Context, msg []byte) error
}

type QueuePair struct {
	Name   string
	Input  Queue
	Output Queue
}
