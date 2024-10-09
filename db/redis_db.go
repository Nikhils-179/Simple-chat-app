package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	cacheClient *redis.Client
	ctx         = context.Background()
)

func createCacheClient() *redis.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cacheClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("Redis_Endpoint"),
		Password: os.Getenv("Redis_Password"),
		DB:       0,
	})

	_, err = cacheClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("here")
		log.Fatalf("Error creating cache client: %v", err)
	}

	return cacheClient
}

func ConnectToCacheDB() *redis.Client {
	if cacheClient == nil {
		createCacheClient()
	}
	return cacheClient
}

func SetCache(ctx context.Context, key string, value string) error {
	return ConnectToCacheDB().Set(ctx, key, value, time.Hour).Err()
}

func GetCache(ctx context.Context, key string) (string, error) {
	return ConnectToCacheDB().Get(ctx, key).Result()
}

func StoreCachedMessage(ctx context.Context, roomID string, message string) error {
	key := fmt.Sprintf("chat:%s", roomID)

	existingMessages, err := GetCachedHistory(ctx, roomID)
	if err != nil && err != redis.Nil {
		return err
	}
	existingMessages = append(existingMessages, message)
	if len(existingMessages) > 50 {
		existingMessages = existingMessages[len(existingMessages)-50:]
	}

	jsonMessages, err := json.Marshal(existingMessages)
	if err != nil {
		return err
	}
	return SetCache(ctx, key, string(jsonMessages))
}

func GetCachedHistory(ctx context.Context, roomID string) ([]string, error) {
	key := fmt.Sprintf("chat:%s", roomID)
	messages, err := GetCache(ctx, key)
	if err != nil {
		return nil, err
	}
	var messageList []string
	err = json.Unmarshal([]byte(messages), &messageList)
	if err != nil {
		return nil, err
	}
	return messageList, nil
}

func FlushHistoryDatabase() error {
	err := ConnectToCacheDB().FlushAll(ctx).Err()
	if err != nil {
		log.Printf("Error flushing history database: %v", err)
		return err
	}
	log.Println("Redis history database flushed successfully")
	return nil
}