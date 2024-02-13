package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"websocket/app/internal/rabbitmq"
)

type client struct {
	UserId  int
	socket  *websocket.Conn
	receive chan []byte
	room    *webSocket
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		c.room.messages <- &rabbitmq.MessageWrapper{
			Message: string(msg),
			UserId:  c.UserId,
		}
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		fmt.Println(msg)
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
