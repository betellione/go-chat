package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients = make(map[*Client]bool)
	mutex   = sync.Mutex{}
	ctx     = context.Background()
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
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

	client := newClient(conn)

	lastMessages, err := getLastMessages(30)
	if err != nil {
		log.Println("Error getting last messages:", err)
	} else {
		for _, msg := range lastMessages {
			client.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}

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

		err = redisClient.Set(ctx, "last_message_from:"+client.ID, msg, 0).Err()
		if err != nil {
			log.Println("Redis Error:", err)
		}

		_, err = db.ExecContext(ctx, "INSERT INTO messages (client_id, message) VALUES ($1, $2)", client.ID, msg)
		if err != nil {
			log.Println("PostgreSQL Error:", err)
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
