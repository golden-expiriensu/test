package rabbitmq

import (
	"context"
	"time"

	"arkis_test/internal/domain/entities"
	"arkis_test/internal/domain/interfaces"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Queue struct {
	amqpConnection *amqp.Connection
	channel        *amqp.Channel
	name           string
}

var _ interfaces.Queue = &Queue{}

func New(connectionURL, queueName string) (interfaces.Queue, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Queue{conn, ch, queueName}, nil
}

func (rmq *Queue) Consume(ctx context.Context) (<-chan entities.Delivery, error) {
	deliveries, err := rmq.channel.ConsumeWithContext(
		ctx,
		rmq.name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	out := make(chan entities.Delivery)

	go func() {
		defer close(out)
		for {
			select {
			case delivery, ok := <-deliveries:
				if !ok {
					zap.L().With(zap.String("queueName", rmq.name)).Info("channel closed")
					return
				}
				out <- entities.Delivery{Body: delivery.Body}
			case <-ctx.Done():
				zap.L().With(zap.String("queueName", rmq.name)).Info("no longer consuming from queue")
				return
			}
		}
	}()

	return out, nil
}

func (rmq *Queue) Publish(ctx context.Context, msg []byte) error {
	data := amqp.Publishing{
		DeliveryMode:    amqp.Transient,
		Timestamp:       time.Now(),
		Body:            msg,
		ContentEncoding: "application/json",
	}

	return rmq.channel.PublishWithContext(ctx, "", rmq.name, true, false, data)
}

func (rmq *Queue) Close() error {
	if err := rmq.channel.Close(); err != nil {
		return err
	}
	return rmq.amqpConnection.Close()
}
