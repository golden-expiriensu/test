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
	Input  Queue
	Output Queue
}

func main() {
	ctx := context.Background()
	rabbitmqUrl := viper.GetString("RABBITMQ_URL")
	var queues []*QueuePair

	inputQueueA, err := queue.New(rabbitmqUrl, "input-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create input queue A")
	}

	outputQueueA, err := queue.New(rabbitmqUrl, "output-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create output queue A")
	}

	queues = append(queues, &QueuePair{
		Input:  inputQueueA,
		Output: outputQueueA,
	})

	inputQueueB, err := queue.New(rabbitmqUrl, "input-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create input queue A")
	}

	outputQueueB, err := queue.New(rabbitmqUrl, "output-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create output queue A")
	}

	queues = append(queues, &QueuePair{
		Input:  inputQueueB,
		Output: outputQueueB,
	})

	log.Info("Application is ready to run")

	wg := sync.WaitGroup{}
	var errs []error
	for i := range 2 {
		log.Infof("Add process %d", i)
		wg.Add(1)
		go func(i int) {
			if err := processor.New(queues[i].Input, queues[i].Output, database.D{}).Run(ctx); err != nil {
				errs = append(errs, err)
			}
			wg.Done()
		}(i)
	}

	queues[0].Input.Publish(ctx, []byte("this is a test"))

	wg.Wait()
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
