package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"websocket/app/internal/rabbitmq"
)

type room struct {
	clients  map[string]*client
	join     chan *client
	leave    chan *client
	messages chan *rabbitmq.MessageWrapper
}

func NewRoom() *room {
	return &room{
		messages: make(chan *rabbitmq.MessageWrapper),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[string]*client),
	}
}

func (r *room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client.UserId] = client
		case client := <-r.leave:
			delete(r.clients, client.UserId)
			close(client.receive)
		case msg := <-r.messages:
			for userId, client := range r.clients {
				if userId == msg.UserId {
					client.receive <- msg.Message
				}
			}
		}
	}
}

func (r *room) AddMessage(msg *rabbitmq.MessageWrapper) {
	r.messages <- msg
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 1256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
	}

	client := &client{
		UserId:  req.URL.Query().Get("userId"),
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}

	r.join <- client
	defer func() {
		r.leave <- client
	}()

	go client.write()
	client.read()
}
