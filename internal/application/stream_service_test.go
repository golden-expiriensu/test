package application

import (
	"context"
	"testing"

	"arkis_test/internal/config"
	"arkis_test/internal/domain/entities"
	"arkis_test/internal/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Consume(ctx context.Context) (<-chan entities.Delivery, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan entities.Delivery), args.Error(1)
}

func (m *MockQueue) Publish(ctx context.Context, msg []byte) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *MockQueue) Close() error {
	return m.Called().Error(0)
}

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Get(key []byte) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		setupMocks func() *StreamService
		expectErr  error
	}{
		{
			name:       "stream already exists",
			streamName: "existing-stream",
			setupMocks: func() *StreamService {
				cfg := &config.Config{RabbitmqUrl: "amqp://localhost:5672"}
				db := &MockDatabase{}
				service := NewStreamService(cfg, db)
				mockInputQueue := &MockQueue{}
				mockOutputQueue := &MockQueue{}

				service.existingStreams.Store("existing-stream", &stream{
					input:          mockInputQueue,
					output:         mockOutputQueue,
					outputReceiver: make(<-chan entities.Delivery),
				})

				return service
			},
			expectErr: errors.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupMocks()
			err := service.Create(context.Background(), tt.streamName)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		message    string
		setupMocks func() *StreamService
		expectErr  error
	}{
		{
			name:       "success",
			streamName: "test-stream",
			message:    "hello",
			setupMocks: func() *StreamService {
				cfg := &config.Config{RabbitmqUrl: "amqp://localhost:5672"}
				db := &MockDatabase{}
				service := NewStreamService(cfg, db)
				mockQueue := &MockQueue{}

				mockQueue.On("Publish", mock.Anything, []byte("hello")).Return(nil)

				service.existingStreams.Store("test-stream", &stream{
					input: mockQueue,
				})

				return service
			},
			expectErr: nil,
		},
		{
			name:       "stream not found",
			streamName: "nonexistent-stream",
			message:    "hello",
			setupMocks: func() *StreamService {
				cfg := &config.Config{RabbitmqUrl: "amqp://localhost:5672"}
				db := &MockDatabase{}
				return NewStreamService(cfg, db)
			},
			expectErr: errors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupMocks()
			err := service.SendMessage(context.Background(), tt.streamName, tt.message)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestReceiveMessage(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		setupMocks func() (*StreamService, <-chan entities.Delivery)
		expectMsg  string
		expectErr  error
	}{
		{
			name:       "success",
			streamName: "test-stream",
			setupMocks: func() (*StreamService, <-chan entities.Delivery) {
				cfg := &config.Config{RabbitmqUrl: "amqp://localhost:5672"}
				db := &MockDatabase{}
				service := NewStreamService(cfg, db)
				mockQueue := &MockQueue{}

				deliveries := make(chan entities.Delivery, 1)
				deliveries <- entities.Delivery{Body: []byte("test-message")}
				close(deliveries)

				mockQueue.On("Consume", mock.Anything).Return(deliveries, nil)
				db.On("Get", []byte("test-message")).Return("test-result", nil)

				service.existingStreams.Store("test-stream", &stream{
					outputReceiver: deliveries,
				})

				return service, deliveries
			},
			expectMsg: "test-result",
			expectErr: nil,
		},
		{
			name:       "stream not found",
			streamName: "nonexistent-stream",
			setupMocks: func() (*StreamService, <-chan entities.Delivery) {
				cfg := &config.Config{RabbitmqUrl: "amqp://localhost:5672"}
				db := &MockDatabase{}
				return NewStreamService(cfg, db), nil
			},
			expectMsg: "",
			expectErr: errors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _ := tt.setupMocks()
			msg, err := service.ReceiveMessage(tt.streamName)
			assert.Equal(t, tt.expectMsg, msg)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}
