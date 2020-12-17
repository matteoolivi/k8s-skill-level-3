package rabbitmq

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

const (
	// TODO: Make these configurable
	host = "rabbitmq"
	port = "5672"

	userEnvVar     = "RABBITMQ_USER"
	passwordEnvVar = "RABBITMQ_PASSWORD"
	queueEnvVar    = "RABBITMQ_QUEUE"

	amqpProto = "amqp"

	uriListMIMEType = "text/uri-list"
)

func EnqueueURL(imageName, imageURL string) error {
	queueName, brokerChannel, cleanUp, err := setUpQueue()
	if err != nil {
		return err
	}
	defer cleanUp()

	err = brokerChannel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			// Use URI list type because there's no type for single URIs.
			ContentType: uriListMIMEType,
			Body:        []byte(imageURL),
		})
	if err != nil {
		return err
	}
	log.Printf("Published URL %s for image %s in RabbitMQ queue %s", imageURL, imageName, queueName)

	return nil
}

// GetURLsChannel declares a queue using the environment variables defined in this package.
// Return args:
//  - channel to receive messages from the queue.
//  - clean up function to invoke when done using the queue. Callers should defer invocation of this
// 	function immediately after a successful invocation (i.e. no errors returned) of GetURLsChannel.
func GetURLsChannel() (<-chan amqp.Delivery, func(), error) {
	queueName, brokerChannel, cleanUp, err := setUpQueue()
	if err != nil {
		return nil, nil, err
	}

	// TODO: Make first two parameters (prefetch count and size) configurable.
	err = brokerChannel.Qos(1, 0, false)
	if err != nil {
		cleanUp()
		return nil, nil, err
	}

	// TODO: Pass in a meaningful consumer name (second argument).
	queueChannel, err := brokerChannel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Acquired channel to read messages from RabbitMQ queue %s", queueName)

	return queueChannel, cleanUp, nil
}

// setUpQueue declares a queue using the environment variables defined in this package.
// Return args:
//  - name of the queue.
//  - channel to send/receive messages from the queue.
//  - clean up function to invoke when done using the queue. Callers should defer invocation of this
// 	function immediately after a successful invocation (i.e. no errors returned) of SetUpQueue.
//  - errors that occurred while setting up the queue.
func setUpQueue() (string, *amqp.Channel, func(), error) {
	brokerURL := buildAMQPurl()
	conn, err := amqp.Dial(brokerURL)
	if err != nil {
		return "", nil, nil, err
	}
	log.Printf("Connected to RabbitMQ broker at URL %s", brokerURL)

	ch, err := conn.Channel()
	if err != nil {
		if err := conn.Close(); err != nil {
			log.Printf("Failed closing connection to RabbitMQ broker at URL %s: %s", brokerURL, err)
		}
		return "", nil, nil, err
	}
	log.Printf("Acquired channel to RabbitMQ broker at URL %s", brokerURL)

	cleanUp := func() {
		if err := ch.Close(); err != nil {
			log.Printf("Failed closing channel to RabbitMQ broker at URL %s: %s", brokerURL, err)
		}
		if err := conn.Close(); err != nil {
			log.Printf("Failed closing connection to RabbitMQ broker at URL %s: %s", brokerURL, err)
		}
	}

	queue, err := ch.QueueDeclare(os.Getenv(queueEnvVar), true, false, false, false, nil)
	if err != nil {
		cleanUp()
		return "", nil, nil, err
	}
	log.Printf("Successfully declared RabbitMQ queue %s at broker with URL %s",
		queue.Name,
		brokerURL)

	return queue.Name, ch, cleanUp, nil
}

func buildAMQPurl() string {
	return amqpProto +
		"://" +
		os.Getenv(userEnvVar) +
		":" +
		os.Getenv(passwordEnvVar) +
		"@" +
		host +
		":" +
		port +
		"/"
}
