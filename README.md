## Description
![untitled](https://github.com/user-attachments/assets/e0438042-1589-4dd1-9d07-c146347adcd1)

This project is a simple chat application built over a WebSocket connection using Go and HTML/JavaScript. It allows multiple clients to connect, send messages, and see the messages from other connected clients in real-time.

## Approach

1. **Create a Client Struct**: The client struct represents a single connected user in the chat room. It includes the WebSocket connection, a channel for receiving messages, and a reference to the room.

2. **Create a Room Struct**: The room struct represents the chat room itself, managing the connected clients, handling client joining and leaving, and forwarding messages to all clients.

3. **WebSocket Handling**: The server upgrades HTTP connections to WebSocket connections and handles sending and receiving messages through the WebSocket connection.

4. **HTTP Server Setup**: An HTTP server serves the chat interface and handles WebSocket connections. The server listens for incoming requests and manages client interactions through the WebSocket connections.

## Usage

1. **Run the Server**: To start the server, execute the following command in your terminal:

    ```bash
    go run main.go
    ```

    This will start the server on the default address `:8080`.

2. **Access the Chat Interface**: Open a web browser and navigate to `http://localhost:8080` to access the chat interface.

3. **Start Chatting**: You can now start sending messages. Open the chat interface in multiple browser windows to simulate multiple clients.

## Project Structure

```bash
.
├── client.go
├── go.mod
├── go.sum
├── main.go
├── room.go
└── templates
    └── chat.html
```

## Dependencies

- **Go Packages**:
  - `github.com/gorilla/websocket`: Used for handling WebSocket connections.
