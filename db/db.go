package db

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func createClient() *redis.Client {
	ctx := context.Background()

	client = redis.NewClient(&redis.Options{
		Addr:     "redis-13538.c282.east-us-mz.azure.redns.redis-cloud.com:13538",
		Password: "Niku@5632",
		DB:       0,
		// Optionally configure TLS if needed
		// TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})

	_, err := client.Ping(ctx).Result()
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
