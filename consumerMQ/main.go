package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

type Message struct {
	Content string `json:"content"`
}

func main() {
	conn := connectToRabbitMQ()
	defer conn.Close()

	ch := openChannel(conn)
	defer ch.Close()

	msgs := consumeMessages(ch)

	processMessages(msgs)
}

func connectToRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	return conn
}

func openChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func consumeMessages(ch *amqp.Channel) <-chan amqp.Delivery {
	msgs, err := ch.Consume("testQueue", "", false, false, false, false, nil)
	failOnError(err, "Failed to register a consumer")
	return msgs
}

func processMessages(msgs <-chan amqp.Delivery) {
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			processMessage(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func processMessage(d amqp.Delivery) {
	log.Printf(" [x] Sending message to receiver")

	var msg Message
	if !json.Valid(d.Body) {
		log.Printf(" [x] Received invalid JSON: %s", d.Body)
		d.Nack(false, true)
		return
	}

	err := json.Unmarshal(d.Body, &msg)
	failOnError(err, "Failed to decode message body")

	body, err := json.Marshal(msg)
	failOnError(err, "Failed to encode message body")

	resp, err := http.Post("http://localhost:8081/receiver/test", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode >= 400 {
		log.Printf(" [x] Failed to send message to receiver, requeueing: %s", d.Body)
		d.Nack(false, true)
		time.Sleep(10 * time.Second)

	} else {
		log.Printf(" [x] Message sent to receiver: %s", d.Body)
		d.Ack(false)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
