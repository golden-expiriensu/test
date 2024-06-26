package publish

import (
	"arkis_test/queue"
	"arkis_test/types"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateUniqueQueueName(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func TestPublishMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("PublishTestQueue")
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

	publishUseCase := NewPublishUseCase()
	originalMessage := "hello"
	err = publishUseCase.PublishMessage(ctx, queuePair, originalMessage)
	assert.NoError(t, err)

	receivedData, err := queuePair.Output.Consume(ctx)
	assert.NoError(t, err)

	select {
	case delivery := <-receivedData:
		msg := string(delivery.Body)
		assert.Equal(t, originalMessage, msg)
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}

func TestPublishMessage_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"

	queueName := generateUniqueQueueName("PublishErrorHandlingTestQueue")
	queueB, err := queue.New(rabbitmqURL, queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queueB)

	queuePair := &types.QueuePair{
		Name: queueName,
		// Input: queueA,
		Output: queueB,
	}

	msg := "hello"
	publishUseCase := NewPublishUseCase()
	err = publishUseCase.PublishMessage(ctx, queuePair, msg)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("failed to publish '%s' to input-%s, error: queue or input channel is nil", msg, queueName), err.Error())
}
