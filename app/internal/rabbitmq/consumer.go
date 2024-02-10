package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

type RabbitMq struct {
	Messages chan *MessageWrapper
}

func ReadFromRabbitMq() {
	fmt.Println("Consume Application")

	conn, err := amqp.Dial("amqp://guest:guest@0.0.0.0:7079")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer ch.Close()

	messages, err := ch.Consume(
		"websocket",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	rabbitMq := &RabbitMq{
		Messages: make(chan *MessageWrapper),
	}

	forever := make(chan bool)

	go func() {
		var message *MessageWrapper

		for d := range messages {
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				fmt.Println(err)
			}
			rabbitMq.Messages <- message
		}
	}()

	fmt.Println("Successfully connected to RabbitMq instance")
	fmt.Println(" [*] - waiting for messages")

	<-forever
}
