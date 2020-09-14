package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)
type RabbitInterface interface {
	GetQueueName() string
	PostMessage(targetQueue string, msg HubMessage)
}
var RabbitService RabbitInterface
const queueName = "single-unity-rabbit-q"

type rabbitService struct {
	*amqp.Connection
	queueName string
}

func (r *rabbitService) GetQueueName() string {
	return r.queueName
}

func (r *rabbitService) PostMessage(targetQueue string, msg HubMessage) {
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		failOnError(err, "could not marshall message for rmq")
	}

	ch, err := r.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.Publish(
		"",
		targetQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonMessage,
		})
	if err != nil {
		failOnError(err, "failed to publish message")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func init() {
	var (
		//TODO: read values from env
		user = "user"
		host     = "unity-msg-svc-rabbitmq"
		port     = "5672"
		password = "password"
	)
	conn, err := amqp.Dial("amqp://"+ user + ":" + password + "@" + host + ":" + port + "/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	//queueName, _ := uuid.NewRandom()
	//for the purpose of testing/debugging using only 1 instance, hardcode the queue name.
	_, err = ch.QueueDeclare(
		queueName, // name
		true,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	messageChannel, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to consume")
	go func() {
		for d:= range messageChannel {
			hubMessage := &HubMessage{}
			err := json.Unmarshal(d.Body, hubMessage)
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}
			MessageHub.QueueMessages <- *hubMessage
			fmt.Printf("Received message from rabbit: %+v \n", hubMessage)
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			}
		}
	}()

	RabbitService = &rabbitService{conn, queueName}
}