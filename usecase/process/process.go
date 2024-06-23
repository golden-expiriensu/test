package process

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/types"
	"context"
	"fmt"
)

type ProcessUseCase struct{}

func NewProcessUseCase() *ProcessUseCase {
	return &ProcessUseCase{}
}

func (p *ProcessUseCase) StartQueueProcess(ctx context.Context, q *types.QueuePair) error {
	if err := processor.New(q.Input, q.Output, database.D{}).Run(ctx); err != nil {
		return fmt.Errorf("Failed to start process for queue: %s, error: %s", q.Name, err.Error())
	}

	return nil
}
