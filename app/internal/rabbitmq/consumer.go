package rabbitmq

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"log"
	"time"
	"websocket/app/internal/config"
)

type RabbitClient struct {
	sendConn *amqp.Connection
	recConn  *amqp.Connection
	sendChan *amqp.Channel
	recChan  *amqp.Channel
}

func (rcl *RabbitClient) connect(isRec, reconnect bool) (*amqp.Connection, error) {
	if reconnect {
		if isRec {
			rcl.recConn = nil
		} else {
			rcl.sendConn = nil
		}
	}
	if isRec && rcl.recConn != nil {
		return rcl.recConn, nil
	} else if !isRec && rcl.sendConn != nil {
		return rcl.sendConn, nil
	}

	cnf := config.GetConfig()
	rabbitConfig := cnf.RabbitConfig

	var c string

	if rabbitConfig.Username == "" {
		c = fmt.Sprintf("amqp://%s:%s/", rabbitConfig.Host, rabbitConfig.Port)
	} else {
		c = fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitConfig.Username, rabbitConfig.Password, rabbitConfig.Host, rabbitConfig.Port)
	}

	conn, err := amqp.Dial(c)
	if err != nil {
		log.Printf("\r\n--- could not create a conection ---\r\n")
		time.Sleep(1 * time.Second)
		return nil, err
	}
	if isRec {
		rcl.recConn = conn
		return rcl.recConn, nil
	} else {
		rcl.sendConn = conn
		return rcl.sendConn, nil
	}
}

func (rcl *RabbitClient) channel(isRec, recreate bool) (*amqp.Channel, error) {
	if recreate {
		if isRec {
			rcl.recChan = nil
		} else {
			rcl.sendChan = nil
		}
	}
	if isRec && rcl.recConn == nil {
		rcl.recChan = nil
	}
	if !isRec && rcl.sendConn == nil {
		rcl.recChan = nil
	}
	if isRec && rcl.recChan != nil {
		return rcl.recChan, nil
	} else if !isRec && rcl.sendChan != nil {
		return rcl.sendChan, nil
	}
	for {
		_, err := rcl.connect(isRec, recreate)
		if err == nil {
			break
		}
	}
	var err error
	if isRec {
		rcl.recChan, err = rcl.recConn.Channel()
	} else {
		rcl.sendChan, err = rcl.sendConn.Channel()
	}
	if err != nil {
		log.Println("--- could not create channel ---")
		time.Sleep(1 * time.Second)
		return nil, err
	}
	if isRec {
		return rcl.recChan, err
	} else {
		return rcl.sendChan, err
	}
}

func (rcl *RabbitClient) Consume(queue string, handler func(data []byte) error) {
	for {
		for {
			_, err := rcl.channel(true, true)
			if err == nil {
				break
			}
		}
		log.Printf("--- connected to consume '%s' ---\r\n", queue)
		q, err := rcl.recChan.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			amqp.Table{"x-queue-mode": "lazy"},
		)
		if err != nil {
			log.Println("--- failed to declare a queue, trying to reconnect ---")
			continue
		}
		connClose := rcl.recConn.NotifyClose(make(chan *amqp.Error))
		connBlocked := rcl.recConn.NotifyBlocked(make(chan amqp.Blocking))
		chClose := rcl.recChan.NotifyClose(make(chan *amqp.Error))

		messages, err := rcl.recChan.Consume(
			q.Name,
			uuid.NewV4().String(),
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			log.Println("--- failed to consume from queue, trying again ---")
			continue
		}

		shouldBreak := false
		for {
			fmt.Println("messages")
			if shouldBreak {
				break
			}
			fmt.Println("messages")
			select {
			case _ = <-connBlocked:
				log.Println("--- connection blocked ---")
				fmt.Println("--- connection blocked ---")
				shouldBreak = true
				break
			case err = <-connClose:
				log.Println("--- connection closed ---")
				fmt.Println("--- connection closed ---")
				shouldBreak = true
				break
			case err = <-chClose:
				log.Println("--- channel closed ---")
				fmt.Println("--- channel closed ---")
				shouldBreak = true
				break
			case d := <-messages:
				err := handler(d.Body)
				fmt.Println(err)
				if err != nil {
					_ = d.Ack(false)
					break
				}
				_ = d.Ack(true)
			}
		}
	}
}

func (rcl *RabbitClient) Publish(n string, b []byte) {
	r := false
	for {
		for {
			_, err := rcl.channel(false, r)
			if err == nil {
				break
			}
		}
		q, err := rcl.sendChan.QueueDeclare(
			n,
			true,
			false,
			false,
			false,
			amqp.Table{"x-queue-mode": "lazy"},
		)
		if err != nil {
			log.Println("--- failed to declare a queue, trying to resend ---")
			r = true
			continue
		}
		err = rcl.sendChan.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				MessageId:    uuid.NewV4().String(),
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         b,
			})
		if err != nil {
			log.Println("--- failed to publish to queue, trying to resend ---")
			r = true
			continue
		}
		break
	}
}
