package publish

import (
	"arkis_test/types"
	"context"
	"fmt"
)

type PublishUseCase struct{}

func NewPublishUseCase() *PublishUseCase {
	return &PublishUseCase{}
}

func (p *PublishUseCase) PublishMessage(ctx context.Context, queue *types.QueuePair, msg string) error {
	if err := queue.Input.Publish(ctx, []byte(msg)); err != nil {
		return fmt.Errorf("failed to publish '%s' to input-%s, error: %s", msg, queue.Name, err.Error())
	}

	return nil
}