package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"websocket/app/internal/rabbitmq"
)

type webSocket struct {
	clients  map[int]*client
	join     chan *client
	leave    chan *client
	messages chan *rabbitmq.MessageWrapper
}

func NewWebSocketServer() *webSocket {
	return &webSocket{
		messages: make(chan *rabbitmq.MessageWrapper),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[int]*client),
	}
}

func (r *webSocket) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client.UserId] = client
		case client := <-r.leave:
			delete(r.clients, client.UserId)
			close(client.receive)
		case msg := <-r.messages:
			for userId, client := range r.clients {
				fmt.Println(userId, msg.UserId)
				if userId == msg.UserId {
					client.receive <- []byte(msg.Message)
				}
			}
		}
	}
}

func (r *webSocket) ListenRabbitQueue() {
	fmt.Println("Start listening a RabbitMq queue")

	var handler = func(data []byte) error {
		var message *rabbitmq.MessageWrapper

		err := json.Unmarshal(data, &message)
		if err != nil {
			return err
		}

		r.messages <- message
		return nil
	}
	rabbitmq.ConsumeFromRabbitMq("websocket", handler)

	//var rabbitClient rabbitmq.RabbitClient
	//rabbitClient.Consume("websocket", handler)
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 1256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *webSocket) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
	}

	userId, err := strconv.Atoi(req.URL.Query().Get("userId"))
	if err != nil {
		fmt.Println(err)
	}

	client := &client{
		UserId:  userId,
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
