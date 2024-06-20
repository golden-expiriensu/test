package main

import (
	"arkis_test/database"
	"arkis_test/processor"
	"arkis_test/queue"
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	ctx := context.Background()

	inputQueue, err := queue.New(viper.GetString("RABBITMQ_URL"), "input-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create input queue")
	}

	outputQueue, err := queue.New(viper.GetString("RABBITMQ_URL"), "output-A")
	if err != nil {
		log.WithError(err).Panic("Cannot create output queue")
	}

	log.Info("Application is ready to run")

	processor.New(inputQueue, outputQueue, database.D{}).Run(ctx)
}

func init() {
	envFile := flag.String("env", ".env", "env file")
	flag.Parse()
	viper.SetConfigName(*envFile)
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read .env file, error: %s", err.Error()))
	}
}
