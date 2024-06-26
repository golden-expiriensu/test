package general

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateUniqueQueueName(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func TestCreateQueuePair(t *testing.T) {
	rabbitmqURL := "amqp://guest:guest@localhost:5672/"
	generalUseCase := NewGeneralUseCase(rabbitmqURL)

	queueName := generateUniqueQueueName("GeneralTestQueue")
	queuePair, err := generalUseCase.CreateQueuePair(queueName)
	assert.NoError(t, err)
	assert.NotNil(t, queuePair)
	assert.Equal(t, queueName, queuePair.Name)
	assert.NotNil(t, queuePair.Input)
	assert.NotNil(t, queuePair.Output)
}

func TestCreateQueuePair_InputQueueError(t *testing.T) {
	rabbitmqURL := "invalidURL"
	generalUseCase := NewGeneralUseCase(rabbitmqURL)

	queueName := generateUniqueQueueName("InputQueueErrorTestQueue")
	queuePair, err := generalUseCase.CreateQueuePair(queueName)
	assert.Error(t, err)
	assert.Nil(t, queuePair)
	assert.Contains(t, err.Error(), fmt.Sprintf("failed to create input queue %s", queueName))
}
