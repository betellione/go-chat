package main

import (
	"database/sql"
	"fmt"
	"github.com/spf13/viper"
	"log"
)

var db *sql.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s sslmode=disable",
		viper.GetString("DB_HOST"),
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_DB"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("DB_PORT"))

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	ensureTableExists()
}

func ensureTableExists() {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		client_id VARCHAR(255),
		message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to ensure that table 'messages' exists: %v", err)
	}
	log.Println("Table 'messages' checked/created successfully")
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
