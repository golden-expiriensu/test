package interfaces

import (
	"context"

	"arkis_test/internal/domain/entities"
)

type Queue interface {
	Consume(ctx context.Context) (<-chan entities.Delivery, error)
	Publish(ctx context.Context, msg []byte) error
	Close() error
}
