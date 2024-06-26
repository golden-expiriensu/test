package consume

import (
	"arkis_test/queue"
	"arkis_test/types"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateUniqueQueueName(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func TestConsumeMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("ConsumeTestQueue")
	queueA, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueA)

	queueB, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueB)

	queuePair := &types.QueuePair{
		Name:   queueName,
		Input:  queueA,
		Output: queueB,
	}

	originalMessage := "hello"
	err = queueA.Publish(ctx, []byte(originalMessage))
	assert.NoError(t, err)

	consumeUseCase := NewConsumeUseCase()
	err = consumeUseCase.ConsumeMessage(ctx, queuePair, func(msg []byte) error {
		receivedMsg := string(msg)
		assert.Equal(t, originalMessage, receivedMsg)

		return nil
	})

	assert.NoError(t, err)
}

func TestConsumeMessage_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("ErrorHandlingTestQueue")
	queueA, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueA)

	queueB, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueB)

	queuePair := &types.QueuePair{
		Name:   queueName,
		Input:  queueA,
		Output: queueB,
	}

	originalMessage := "hello"
	err = queueA.Publish(ctx, []byte(originalMessage))
	assert.NoError(t, err)

	consumeUseCase := NewConsumeUseCase()
	err = consumeUseCase.ConsumeMessage(ctx, queuePair, func(msg []byte) error {
		return fmt.Errorf("simulated error")
	})

	assert.Error(t, err)
}

func TestConsumeMessage_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("TimeoutTestQueue")
	queueA, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueA)

	queueB, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueB)

	queuePair := &types.QueuePair{
		Name:   queueName,
		Input:  queueA,
		Output: queueB,
	}

	time.Sleep(1 * time.Second)
	consumeUseCase := NewConsumeUseCase()
	err = consumeUseCase.ConsumeMessage(ctx, queuePair, func(msg []byte) error {
		receivedMsg := string(msg)
		assert.Equal(t, "", receivedMsg)
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("failed to consume messages from output-%s, error: context deadline exceeded", queueName), err.Error())
}

func TestConsumeMessage_InvalidMessageFormat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("InvalidFormatTestQueue")
	queueA, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueA)

	queueB, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueB)

	queuePair := &types.QueuePair{
		Name:   queueName,
		Input:  queueA,
		Output: queueB,
	}

	invalidMessage := `{"invalidJson":`
	err = queueA.Publish(ctx, []byte(invalidMessage))
	assert.NoError(t, err)

	consumeUseCase := NewConsumeUseCase()
	err = consumeUseCase.ConsumeMessage(ctx, queuePair, func(msg []byte) error {
		var jsonData map[string]interface{}
		err := json.Unmarshal(msg, &jsonData)
		assert.Error(t, err)
		return nil
	})

	assert.NoError(t, err)
}
