package rabbitmq

import (
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
)
type RabbitInterface interface {
	GetQueueName() string
}
var RabbitService RabbitInterface

type rabbitService struct {
	*amqp.Connection
	queueName string
}

func (r *rabbitService) GetQueueName() string {
	return r.queueName
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func init() {
	var (
		//TODO: read values from env
		user = "guest"
		host     = "localhost"
		port     = "5672"
		password = "guest"
	)
	conn, err := amqp.Dial("amqp://"+ user + ":" + password + "@" + host + ":" + port + "/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	queueName, _ := uuid.NewRandom()
	_, err = ch.QueueDeclare(
		queueName.String(), // name
		false,              // durable
		true,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")
	//body := "Hello World!"
	//err = ch.Publish(
	//	"",     // exchange
	//	q.Name, // routing key
	//	false,  // mandatory
	//	false,  // immediate
	//	amqp.Publishing {
	//		ContentType: "text/plain",
	//		Body:        []byte(body),
	//	})
	failOnError(err, "Failed to publish a message")

	RabbitService = &rabbitService{conn, queueName.String()}

	defer ch.Close()
}