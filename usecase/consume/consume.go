package consume

import (
	"arkis_test/types"
	"context"
	"fmt"
)

type ConsumeUseCase struct{}

func NewConsumeUseCase() *ConsumeUseCase {
	return &ConsumeUseCase{}
}

func (c *ConsumeUseCase) ConsumeMessage(ctx context.Context, queue *types.QueuePair, callback func(msg string) error) error {
	receivedData, err := queue.Output.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to consume messages from output-%s, error: %s", queue.Name, err.Error())
	}

	select {
	case delivery := <-receivedData:
		msg := string(delivery.Body)
		return callback(msg)
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for message from output-%s", queue.Name)
	}
}
