package main

import (
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	InitConfig()
	InitDB()
	InitRedis()
	http.HandleFunc("/ws", HandleConnections)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "chat.html")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
