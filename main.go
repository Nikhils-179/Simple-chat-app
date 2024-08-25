package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"Nikhils-179/simple-chat-app/db"

	"github.com/gorilla/websocket"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

type client struct {
	socket   *websocket.Conn
	recieve  chan string
	room     *room
	username string
}

type room struct {
	clients        map[*client]bool
	join           chan *client
	leave          chan *client
	forward        chan string
	messageHistory []string
}

func newRoom() *room {
	return &room{
		clients: make(map[*client]bool),
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan string),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.recieve)
		case msg := <-r.forward:
			for client := range r.clients {
				client.recieve <- msg
			}
			// Store message in Redis
			err := db.StoreMessage("room1", msg)
			if err != nil {
				log.Printf("Error storing message: %v", err)
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 1024
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: messageBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	client := &client{
		socket:   socket,
		recieve:  make(chan string, messageBufferSize),
		room:     r,
		username: username,
	}
	r.join <- client
	defer func() {
		r.leave <- client
	}()
	go client.write()
	client.read()
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		if string(msg) == `{"action":"getHistory"}` {
			// Retrieve chat history from Redis
			history, err := db.GetChatHistory("room1")
			if err != nil {
				log.Printf("Error retrieving chat history: %v", err)
				continue
			}
			historyMessage := fmt.Sprintf("%s: %s", c.username, "Chat History:\n"+strings.Join(history, "\n"))
			c.socket.WriteMessage(websocket.TextMessage, []byte(historyMessage))
		} else {
			message := fmt.Sprintf("%s: %s", c.username, string(msg))
			c.room.forward <- message
		}
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.recieve {
		if err := c.socket.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			break
		}
	}
}

func main() {
	client := db.ConnectToDB()
	fmt.Println("Connected to Redis:", client)
	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse()

	r := newRoom()

	// Handle static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("templates"))))

	// Handle template rendering
	http.Handle("/", &templateHandler{filename: "chat.html"})
	// Handle WebSocket connections
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Listen and serve:", err)
	}
}
