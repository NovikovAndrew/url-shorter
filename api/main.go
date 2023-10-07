package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"url-shorter/database"
	"url-shorter/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	db, err := strconv.Atoi(os.Getenv("APP_QUOTA"))
	if err != nil {
		log.Fatalln(err)
	}

	redisOptions := database.RedisOptions{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASSWORD"),
		DB:       db,
	}

	addr := fmt.Sprintf("%s:%s", os.Getenv("DOMAIN"), os.Getenv("APP_PORT"))

	redisClient := database.NewRedisClient(redisOptions)
	server := routes.NewServer(redisClient, addr)
	if err := server.Start(); err != nil {
		log.Fatalln(err)
	}
}
