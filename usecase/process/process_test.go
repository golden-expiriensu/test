package process

import (
	"arkis_test/queue"
	"arkis_test/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateUniqueQueueName(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func TestStartQueueProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"
	queueBaseName := generateUniqueQueueName("ProcessTestQueue")

	inputQueue, err := queue.New(rabbitmqURL, fmt.Sprintf("input-%s", queueBaseName))
	assert.NoError(t, err)
	outputQueue, err := queue.New(rabbitmqURL, fmt.Sprintf("output-%s", queueBaseName))
	assert.NoError(t, err)

	queuePair := &types.QueuePair{
		Name:   queueBaseName,
		Input:  inputQueue,
		Output: outputQueue,
	}

	processUseCase := NewProcessUseCase()

	processDone := make(chan error, 1)
	go func() {
		processDone <- processUseCase.StartQueueProcess(ctx, queuePair)
	}()

	testMessage := "hello"
	err = inputQueue.Publish(ctx, []byte(testMessage))
	assert.NoError(t, err)

	receivedData, err := outputQueue.Consume(ctx)
	assert.NoError(t, err)

	select {
	case delivery := <-receivedData:
		var result map[string]string
		err = json.Unmarshal(delivery.Body, &result)
		assert.NoError(t, err)
		assert.Equal(t, testMessage, result["result"])

		cancel()
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for processed message")
	}

	select {
	case err := <-processDone:
		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				// Expected error --> test is passed
				assert.ErrorContains(t, err, "context canceled")
			} else {
				t.Fatalf("Process ended with unexpected error: %v", err)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Process did not end in time")
	}
}

func TestStartQueueProcess_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rabbitmqURL := "amqp://guest:guest@localhost:5672/"
	queueName := generateUniqueQueueName("ProcessErrorHandlingTestQueue")

	queueA, err := queue.New(rabbitmqURL, fmt.Sprintf("input-%s", queueName))
	assert.NoError(t, err)
	assert.NotNil(t, queueA)

	queuePair := &types.QueuePair{
		Name:  queueName,
		Input: queueA,
	}

	processUseCase := NewProcessUseCase()
	err = processUseCase.StartQueueProcess(ctx, queuePair)
	assert.Error(t, err)
	assert.ErrorContains(t, err, fmt.Sprintf("Failed to start process for queue: %s", queueName))
}
