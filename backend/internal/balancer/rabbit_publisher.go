package balancer

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	url   string
	queue string
}

func NewRabbitPublisher(url, queue string) *RabbitPublisher {
	return &RabbitPublisher{url: url, queue: queue}
}

func (p *RabbitPublisher) PublishAppealNeedsDistribution(appealID, teamID int) error {
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		p.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(RabbitEvent{
		Name: EventAppealNeedsDistribution,
		Payload: RabbitEventBody{
			AppealID: appealID,
			TeamID:   teamID,
		},
	})
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
