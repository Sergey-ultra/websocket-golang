package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
	"websocket/app/internal/rabbitmq"
	"websocket/app/internal/ws"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("app", "templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	start()
}

func start() {
	r := ws.NewRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/ws", r)

	go r.Run()

	var addr = flag.String("addr", ":8091", "The address of application")
	flag.Parse()
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	go rabbitmq.ReadFromRabbitMq()
}
