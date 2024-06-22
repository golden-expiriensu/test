package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Queue interface {
	Consume(ctx context.Context) (<-chan queue.Delivery, error)
	Publish(ctx context.Context, msg []byte) error
}

type QueuePair struct {
	Name   string
	Input  Queue
	Output Queue
}

func main() {
	ctx := context.Background()

	rabbitmqURL := viper.GetString("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Panic("RABBITMQ_URL is not set")
	}

	queueNames := []string{"A", "B"}
	queues := make([]*QueuePair, len(queueNames))

	for i, name := range queueNames {
		queuePair, err := createQueuePair(rabbitmqURL, name)
		if err != nil {
			log.WithError(err).Panic(err)
		}
		queues[i] = queuePair
	}

	log.Info("Application is ready to run")

	wg := sync.WaitGroup{}
	var errs []error
	var mu sync.Mutex

	for i, q := range queues {
		wg.Add(1)
		go func(q *QueuePair, index int) {
			defer wg.Done()
			if err := processor.New(q.Input, q.Output, database.D{}).Run(ctx); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("Processor %d: %v", index, err))
				mu.Unlock()
			}
		}(q, i)
	}

	if err := publishMessages(ctx, queues); err != nil {
		log.WithError(err).Error("Failed to publish messages")
	}

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Panic("One or more processors encountered an error")
	}
}

func publishMessages(ctx context.Context, queues []*QueuePair) error {
	messages := []struct {
		input  string
		output string
	}{
		{"hello", `{"result": "hello"}`},
		{"world", `{"result": "world"}`},
	}

	for i, msg := range messages {
		if err := queues[i].Input.Publish(ctx, []byte(msg.input)); err != nil {
			return fmt.Errorf("failed to publish message to input-%s, error: %s", queues[i].Name, err.Error())
		}
	}

	time.Sleep(2 * time.Second)

	for i, q := range queues {
		deliveries, err := q.Output.Consume(ctx)
		if err != nil {
			return fmt.Errorf("failed to consume messages from output-%s, error: %s", queues[i].Name, err.Error())
		}

		select {
		case delivery := <-deliveries:
			expected := messages[i].output
			if string(delivery.Body) != expected {
				return fmt.Errorf("unexpected message from output-%s, got: %s, expected: %s", queues[i].Name, delivery.Body, expected)
			}
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for message from output-%s", queues[i].Name)
		}
	}

	return nil
}

func createQueuePair(rabbitmqURL, queueName string) (*QueuePair, error) {
	inputQ, err := queue.New(rabbitmqURL, fmt.Sprintf("input-%s", queueName))
	if err != nil {
		return nil, fmt.Errorf("failed to create input queue %s, error: %s", queueName, err.Error())
	}

	outputQ, err := queue.New(rabbitmqURL, fmt.Sprintf("output-%s", queueName))
	if err != nil {
		return nil, fmt.Errorf("failed to create output queue %s, error: %s", queueName, err.Error())
	}

	return &QueuePair{
		Name:   queueName,
		Input:  inputQ,
		Output: outputQ,
	}, nil
}

func init() {
	envFile := flag.String("env", ".env", "env file")
	flag.Parse()
	viper.SetConfigName(*envFile)
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Panic(fmt.Sprintf("failed to read .env file, error: %s", err.Error()))
	}
}
