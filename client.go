package main

import (
	"log"

	"github.com/gorilla/websocket"
)

//client represents a user
type client struct {

	socket *websocket.Conn

	//recieve channel used to recieve message from other clients
	recieve chan []byte

	//room is the space used to define the room client is chatting in 
	room *room 

}

func (c *client) read(){

	defer c.socket.Close() 

	for {
		_ , msg , err := c.socket.ReadMessage()
		if err != nil {
			log.Println("Error reading the message" , err)
			return
		}

		c.room.forward <- msg 
	}
}


func(c *client) write() {
	defer c.socket.Close()

	for msg := range c.recieve {
		err := c.socket.WriteMessage(websocket.TextMessage , msg)
		if err != nil {
			log.Println("Error sending message")
			return
		}
	}
}