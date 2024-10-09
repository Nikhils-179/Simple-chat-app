package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"text/template"

	"Nikhils-179/simple-chat-app/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	ctx      = context.Background()
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) LoadTemplate() {
	t.once.Do(func() {
		var err error
		t.templ, err = template.ParseFiles("templates/" + t.filename)
		if err != nil {
			log.Printf("Error loading template: %v", err)
		}
	})
}

func (t *templateHandler) ServeHTTP(c *gin.Context) {
	t.LoadTemplate()
	if t.templ != nil {
		if err := t.templ.Execute(c.Writer, c.Request); err != nil {
			log.Printf("Error executing template: %v", err)
			c.String(500, "Internal Server Error")
		}
	}
}

type client struct {
	socket   *websocket.Conn
	recieve  chan string
	room     *room
	username string
}

type room struct {
	clients map[*client]bool
	join    chan *client
	leave   chan *client
	forward chan string
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
			// Store message in Redis history DB
			err := db.StoreCachedMessage(ctx, "room1", msg)
			if err != nil {
				log.Printf("Error storing message in cache: %v", err)
			}
			err = db.StoreMessageInMongo("room1", msg)
            if err != nil {
                log.Printf("Error storing message in MongoDB: %v", err)
            }
		}
	}
}

func (r *room) ServeWebSocket(c *gin.Context) {
	username := c.Query("username")
	socket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	client := &client{
		socket:   socket,
		recieve:  make(chan string, 1024),
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
			log.Println("Error reading message:", err)
			return
		}
		if string(msg) == `{"action":"getHistory"}` {
			history, err := db.GetCachedHistory(ctx, "room1")
			if err != nil {
				log.Printf("Error retrieving chat history: %v", err)
				continue
			}
			historyMessage := fmt.Sprintf("%s: %s", c.username, "Chat History:\n"+strings.Join(history, "\n"))
			if err := c.socket.WriteMessage(websocket.TextMessage, []byte(historyMessage)); err != nil {
				log.Println("Error writing message:", err)
			}
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
			log.Println("Error writing message to client:", err)
			break
		}
	}
}

func main() {
	err := db.FlushHistoryDatabase()
	if err != nil {
		log.Fatal("Failed to flush Redis history database:", err)
	}

	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse()

	r := newRoom()
	go r.run()

	router := gin.Default()
	router.Static("/static", "./templates")
	router.GET("/", func(c *gin.Context) {
		handler := &templateHandler{filename: "chat.html"}
		handler.ServeHTTP(c)
	})
	router.GET("/room", r.ServeWebSocket)
	router.GET("/load-test", db.LoadTestHandler)

	server := &http.Server{
		Addr:    *addr,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting web server on", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-done
	log.Println("Shutting down gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
