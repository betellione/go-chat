package main

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

func InitConfig() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	viper.AutomaticEnv()
}
