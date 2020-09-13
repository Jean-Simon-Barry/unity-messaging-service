package rabbitmq

import (
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

	//queueName, _ := uuid.NewRandom()
	//for the purpose of testing/debugging using only 1 instance, hardcode the queue name.
	queueName := "single-unity-rabbit-q"
	_, err = ch.QueueDeclare(
		queueName, // name
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

	RabbitService = &rabbitService{conn, queueName}

	defer ch.Close()
}