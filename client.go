package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	ID   string
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		ID:   uuid.New().String(),
	}
}
