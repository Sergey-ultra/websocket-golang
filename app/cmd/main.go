package main

import (
	"flag"
	"log"
	"net/http"
	"websocket/app/internal/ws"
)

//type templateHandler struct {
//	once     sync.Once
//	filename string
//	templ    *template.Template
//}
//
//func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	t.once.Do(func() {
//		t.templ = template.Must(template.ParseFiles(filepath.Join("app", "templates", t.filename)))
//	})
//	t.templ.Execute(w, r)
//}

func main() {
	start()
}

func start() {
	wsServer := ws.NewWebSocketServer()

	//http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/ws", wsServer)

	go wsServer.Run()
	go wsServer.ListenRabbitQueue()

	var addr = flag.String("addr", ":4000", "The address of application")
	flag.Parse()
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
