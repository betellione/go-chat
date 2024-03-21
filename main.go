package main

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
	ID   string
}

var (
	clients     = make(map[*Client]bool)
	mutex       = sync.Mutex{}
	redisClient *redis.Client
	db          *sql.DB
	ctx         = context.Background()
)

func init() {
	// Подключение к Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Подключение к PostgreSQL
	var err error
	db, err = sql.Open("postgres", "user=chat_admin password=password dbname=chat_db port=6543 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

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

	clientID := uuid.New().String()
	client := &Client{conn: conn, ID: clientID}

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

func getLastMessages(limit int) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT message FROM messages LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
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
