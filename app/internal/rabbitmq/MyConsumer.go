package rabbitmq

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"websocket/app/internal/config"
)

func ConsumeFromRabbitMq(queue string, handler func(data []byte) error) {
	fmt.Println("Consume Application")

	cnf := config.GetConfig()
	rabbitConfig := cnf.RabbitConfig
	c := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitConfig.Username, rabbitConfig.Password, rabbitConfig.Host, rabbitConfig.Port)

	conn, err := amqp.Dial(c)
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
		queue,
		uuid.NewV4().String(),
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

	forever := make(chan bool)

	go func() {
		for d := range messages {
			err := handler(d.Body)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	fmt.Println("Successfully connected to RabbitMq instance")
	fmt.Println(" [*] - waiting for messages")

	<-forever
}
