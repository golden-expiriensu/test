package general

import (
	"arkis_test/queue"
	"arkis_test/types"
	"fmt"
)

type GeneralUseCase struct {
	rabbitmqURL string
}

func NewGeneralUseCase(rabbitmqURL string) *GeneralUseCase {
	return &GeneralUseCase{
		rabbitmqURL: rabbitmqURL,
	}
}

func (g *GeneralUseCase) CreateQueuePair(queueName string) (*types.QueuePair, error) {
	inputQ, err := queue.New(g.rabbitmqURL, fmt.Sprintf("input-%s", queueName))
	if err != nil {
		return nil, fmt.Errorf("failed to create input queue %s, error: %s", queueName, err.Error())
	}

	outputQ, err := queue.New(g.rabbitmqURL, fmt.Sprintf("output-%s", queueName))
	if err != nil {
		return nil, fmt.Errorf("failed to create output queue %s, error: %s", queueName, err.Error())
	}

	return &types.QueuePair{
		Name:   queueName,
		Input:  inputQ,
		Output: outputQ,
	}, nil
}
