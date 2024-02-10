package ws

import (
	"github.com/gorilla/websocket"
	"websocket/app/internal/rabbitmq"
)

type client struct {
	UserId  string
	socket  *websocket.Conn
	receive chan []byte
	room    *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		c.room.messages <- &rabbitmq.MessageWrapper{
			Message: msg,
			UserId:  c.UserId,
		}
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
