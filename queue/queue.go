package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Delivery struct {
	Body    []byte
	handler amqp.Delivery
}

type Queue struct {
	amqpConnection *amqp.Connection
	channel        *amqp.Channel
	name           string
}

func New(connectionURL, queueName string) (*Queue, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
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
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	return &Queue{conn, ch, queueName}, nil
}

func (queue *Queue) Consume(ctx context.Context) (<-chan Delivery, error) {
	deliveries, err := queue.channel.ConsumeWithContext(
		ctx,
		queue.name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("retrieved queued messages: %w", err)
	}

	out := make(chan Delivery)

	go func() {
		for {
			select {
			case delivery := <-deliveries:
				out <- Delivery{delivery.Body, delivery}
			case <-ctx.Done():
				close(out)
				return
			}
		}
	}()

	return out, nil
}

func (queue *Queue) Publish(ctx context.Context, msg []byte) error {
	data := amqp.Publishing{
		DeliveryMode:    amqp.Transient,
		Timestamp:       time.Now(),
		Body:            msg,
		ContentEncoding: "application/json",
	}

	return queue.channel.PublishWithContext(ctx, "", queue.name, true, false, data)
}
