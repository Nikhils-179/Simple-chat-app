package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx = context.Background()


func createClient() *redis.Client {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}else {
		fmt.Println(".env file loaded successfully")
	}


	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("Redis_Endpoint"),
		Password: os.Getenv("Redis_Passowrd"),
		DB:       0,
	})

	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error creating Redis client: %v", err)
	}

	return client
}

func ConnectToDB() *redis.Client {
	if client == nil {
		createClient()
	}
	return client
}

func StoreMessage(roomID string, message string) error {
	ctx := context.Background()
	key := fmt.Sprintf("chat:%s", roomID)
	_, err := client.RPush(ctx, key, message).Result()
	return err
}

func GetChatHistory(roomID string) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("chat:%s", roomID)
	messages, err := client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func FlushDatabase() error {
	client := ConnectToDB()

	err := client.FlushAll(ctx).Err()
	if err != nil {
		log.Printf("Error flushing database: %v", err)
		return err
	}
	log.Println("Redis database flushed successfully")
	return nil
}