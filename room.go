package main

import(
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct{

	//clients hold all the current client in this room
	clients map[*client]bool

	//join channel used to join clients wishing to join the room
	join chan *client

	//leave channel used to help client who wish to leave the room
	leave chan *client

	//forward channel holds the incoming message that should be frwarded to other clients in the room
	forward chan []byte

}

//constructor
func newRoom() *room {
	return &room{
		clients: make(map[*client]bool),
		join: make(chan *client),
		leave: make(chan *client),
		forward: make(chan []byte),
	}
}


func(r *room) run() {
	for {
		select {
		case client := <- r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients,client)
		case msg := <- r.forward:
			for client := range r.clients{
				client.recieve <- msg
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 1024
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize , WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter , req *http.Request){
	socket , err := upgrader.Upgrade(w, req ,nil) 
	if err != nil {
		log.Fatal("Error forwarding to websocket:",err)
		return
	}
	client := &client{
		socket: socket,
		recieve: make(chan []byte , messageBufferSize),
		room: r,
	}
	r.join <-client
	defer func ()  {
		r.leave <- client
	}()
	go client.write()
	client.read()
}