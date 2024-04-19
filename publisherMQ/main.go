package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

type Message struct {
	Content string `json:"content"`
}

func main() {
	conn := connectToRabbitMQ()
	defer conn.Close()

	ch := createChannel(conn)
	defer ch.Close()

	q := declareQueue(ch)
	registerPublishHandler(ch, q)
}

func connectToRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	return conn
}

func createChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func declareQueue(ch *amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare("testQueue", false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")
	return q
}

func registerPublishHandler(ch *amqp.Channel, q amqp.Queue) {
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		failOnError(err, "Failed to decode request body")
		publishMessage(ch, q, msg)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func publishMessage(ch *amqp.Channel, q amqp.Queue, msg Message) {
	body, err := json.Marshal(msg)
	failOnError(err, "Failed to encode message body")

	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	err = ch.Publish("", q.Name, false, false, publishing)
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", msg.Content)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
