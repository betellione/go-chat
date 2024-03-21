package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
}

var (
	clients = make(map[*Client]bool)
	mutex   = sync.Mutex{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	client := &Client{conn: conn}

	mutex.Lock()
	clients[client] = true
	mutex.Unlock()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, client)
			mutex.Unlock()
			break
		}
		broadcastMessage(msgType, msg)
	}
}

func broadcastMessage(msgType int, msg []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		if err := client.conn.WriteMessage(msgType, msg); err != nil {
			log.Println(err)
			delete(clients, client)
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "chat.html")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
