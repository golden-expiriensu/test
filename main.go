package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"flag"
	"fmt"
	"sync"

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

	go func(ctx context.Context, queues []*QueuePair) {
		if err := publishMessages(ctx, queues); err != nil {
			log.WithError(err).Error("Failed to publish messages")
		}
	}(ctx, queues)

	for _, q := range queues {
		go func(ctx context.Context, q *QueuePair) {
			if err := consumeMessages(ctx, q, func(msg string) error {
				log.Infof("received msg: %s, for queue: %s", msg, q.Name)
				return nil
			}); err != nil {
				log.WithError(err).Errorf("Failed to consumer messages for queue %s", q.Name)
			}
		}(ctx, q)
	}

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Panic("One or more processors encountered an error")
	}
}

func consumeMessages(ctx context.Context, queue *QueuePair, callback func(msg string) error) error {
	receivedData, err := queue.Output.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to consume messages from output-%s, error: %s", queue.Name, err.Error())
	}

	select {
	case delivery := <-receivedData:
		msg := string(delivery.Body)
		return callback(msg)
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for message from output-%s", queue.Name)
	}
}

func publishMessages(ctx context.Context, queues []*QueuePair) error {
	messages := []string{"hello", "world"}

	for i, msg := range messages {
		if err := queues[i].Input.Publish(ctx, []byte(msg)); err != nil {
			return fmt.Errorf("failed to publish '%s' to input-%s, error: %s", msg, queues[i].Name, err.Error())
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
