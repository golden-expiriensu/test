package application

import (
	"context"
	"sync"

	"arkis_test/internal/config"
	"arkis_test/internal/domain/entities"
	"arkis_test/internal/domain/interfaces"
	"arkis_test/internal/errors"
	"arkis_test/internal/infrastructure/communication/rabbitmq"

	"go.uber.org/zap"
)

type StreamService struct {
	cfg *config.Config
	db  interfaces.Database

	// existingStreams is a concurrent-safe map that stores the active streams.
	// We use sync.Map instead of a standard map with a mutex because:
	// 1. Streams are created once and then read many times, fitting sync.Map's optimised use case.
	// 2. sync.Map reduces lock contention for frequent read operations, improving performance.
	// 3. It simplifies concurrency handling by avoiding explicit locking logic, enhancing code maintainability.
	existingStreams sync.Map

	wg          sync.WaitGroup
	shutdownCtx context.Context
	cancel      context.CancelFunc
}

// stream consists of an input queue and a receive-only channel of an outputReceiver queue
type stream struct {
	input          interfaces.Queue
	output         interfaces.Queue
	outputReceiver <-chan entities.Delivery
}

func NewStreamService(cfg *config.Config, db interfaces.Database) *StreamService {
	shutdownCtx, cancel := context.WithCancel(context.Background())
	return &StreamService{
		cfg:         cfg,
		db:          db,
		shutdownCtx: shutdownCtx,
		cancel:      cancel,
	}
}

// Create creates a new stream with the given name
func (p *StreamService) Create(ctx context.Context, streamName string) error {
	if _, ok := p.existingStreams.Load(streamName); ok {
		zap.L().With(zap.String("name", streamName)).Error("stream already exists")
		return errors.ErrAlreadyExists
	}

	inputName, outputName := "input-"+streamName, "output-"+streamName

	input, err := rabbitmq.New(p.cfg.RabbitmqUrl, inputName)
	if err != nil {
		zap.L().With(zap.String("input", inputName), zap.Error(err)).Error("failed to create input queue")
		return err
	}

	output, err := rabbitmq.New(p.cfg.RabbitmqUrl, outputName)
	if err != nil {
		zap.L().With(zap.String("outputReceiver", outputName), zap.Error(err)).Error("failed to create outputReceiver queue")
		return err
	}

	outputReceiver, err := output.Consume(p.shutdownCtx)
	if err != nil {
		zap.L().With(zap.String("outputReceiver", outputName), zap.Error(err)).Error("failed to consume from outputReceiver")
		return err
	}

	inputReceiver, err := input.Consume(p.shutdownCtx)
	if err != nil {
		zap.L().With(zap.String("input", inputName), zap.Error(err)).Error("failed to consume from input")
		return err
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		err := p.startTransmission(p.shutdownCtx, inputReceiver, outputName, output)
		if err != nil {
			zap.L().Error("failed to transmit a message", zap.Error(err))
		}
	}()

	p.existingStreams.Store(streamName, &stream{input, output, outputReceiver})
	return nil
}

// SendMessage sends a message to the input queue of the stream with the given name
func (p *StreamService) SendMessage(ctx context.Context, streamName, msg string) error {
	zap.L().With(zap.String("name", streamName)).Info("sending message")

	value, ok := p.existingStreams.Load(streamName)
	if !ok {
		zap.L().With(zap.String("name", streamName)).Error("stream not found")
		return errors.ErrNotFound
	}
	sm := value.(*stream)

	return p.processMessage(ctx, sm.input, []byte(msg))
}

// ReceiveMessage fetches the next message from the outputReceiver channel of the stream with the given name
func (p *StreamService) ReceiveMessage(streamName string) (string, error) {
	value, ok := p.existingStreams.Load(streamName)
	if !ok {
		zap.L().With(zap.String("name", streamName)).Error("stream not found")
		return "", errors.ErrNotFound
	}
	sm := value.(*stream)

	select {
	case delivery, ok := <-sm.outputReceiver:
		if !ok {
			return "", nil
		}

		msg, err := p.db.Get(delivery.Body)
		if err != nil {
			zap.L().With(zap.Error(err)).Error("failed to get message from database")
			return "", err
		}

		return msg, nil
	default:
		return "", nil
	}
}

// Shutdown stops all active streams and transmissions and closes their queues
func (p *StreamService) Shutdown() {
	p.cancel()
	p.existingStreams.Range(func(key, value interface{}) bool {
		sm := value.(*stream)
		if err := sm.input.Close(); err != nil {
			zap.L().Error("failed to close input queue", zap.String("stream", key.(string)), zap.Error(err))
		}
		if err := sm.output.Close(); err != nil {
			zap.L().Error("failed to close output queue", zap.String("stream", key.(string)), zap.Error(err))
		}
		return true
	})
	p.wg.Wait()
}

// startTransmission listens for messages from the input queue and publishes them to the output queue
func (p *StreamService) startTransmission(
	ctx context.Context,
	inputReceiver <-chan entities.Delivery,
	outputName string,
	output interfaces.Queue,
) error {
	for {
		select {
		case <-ctx.Done():
			zap.L().With(zap.String("to", outputName)).Info("stopped transmission")
			return nil
		case delivery, ok := <-inputReceiver:
			if !ok {
				zap.L().With(zap.String("to", outputName)).Info("input receiver channel closed")
				return nil
			}
			if err := p.processMessage(ctx, output, delivery.Body); err != nil {
				zap.L().With(zap.Error(err)).Error("failed to publish message")
				return err
			}
		}
	}
}

// processMessage processes the given message by publishing it to the given queue
// In the future, it can be broken down into separate functions to handle more complex processing logic for each
// individual processing case.
func (p *StreamService) processMessage(ctx context.Context, q interfaces.Queue, msg []byte) error {
	return q.Publish(ctx, msg)
}
