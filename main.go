package main

import (
	"arkis_test/types"
	"arkis_test/usecase/consume"
	"arkis_test/usecase/general"
	"arkis_test/usecase/process"
	"arkis_test/usecase/publish"
	"context"
	"flag"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	ctx := context.Background()

	rabbitmqURL := viper.GetString("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Panic("RABBITMQ_URL is not set")
	}

	queueNames := []string{"A", "B"}
	messages := []string{"hello", "world"}
	queues := make([]*types.QueuePair, len(queueNames))
	generalUseCase := general.NewGeneralUseCase(rabbitmqURL)
	publishUseCase := publish.NewPublishUseCase()
	consumeUseCase := consume.NewConsumeUseCase()
	processUseCase := process.NewProcessUseCase()
	for i, name := range queueNames {
		queuePair, err := generalUseCase.CreateQueuePair(name)
		if err != nil {
			log.WithError(err).Panic(err)
		}
		queues[i] = queuePair
	}

	log.Info("Application is ready to run")
	log.Info("Waiting for messages...")

	wg := sync.WaitGroup{}
	var errs []error
	var mu sync.Mutex

	for i, q := range queues {
		wg.Add(1)
		go func(q *types.QueuePair, index int) {
			defer wg.Done()
			if err := processUseCase.StartQueueProcess(ctx, q); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("Processor %d: %v", index, err))
				mu.Unlock()
			}
		}(q, i)
	}

	for i, q := range queues {

		msg := messages[i]
		go func(ctx context.Context, q *types.QueuePair, msg string) {
			if err := publishUseCase.PublishMessage(ctx, q, msg); err != nil {
				log.WithError(err).Errorf("Failed to publish message: %s, for queue: %s", msg, q.Name)
			}
		}(ctx, q, msg)

		go func(ctx context.Context, q *types.QueuePair) {
			if err := consumeUseCase.ConsumeMessage(ctx, q, func(msg string) error {
				log.Infof("Queue: %s has received this message: %s", q.Name, msg)
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
