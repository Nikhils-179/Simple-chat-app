package db

import (
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func CreateMongoClient() *mongo.Client {
	uri := os.Getenv("MONGODBURI")
	fmt.Println(uri)
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Error Creating Mongo Client %v", err)
	}
	return mongoClient
}

func GetDatabase() *mongo.Database {
	if mongoClient == nil {
		CreateMongoClient()
	}
	return mongoClient.Database("Chat-DB")
}

func StoreMessageInMongo(roomID string , message string) error {
	mongoClient = CreateMongoClient()
	defer mongoClient.Disconnect(ctx)
	collection := mongoClient.Database("Chat-DB").Collection("messages")
	_ , err := collection.InsertOne(ctx , bson.M{"roomID" : roomID , "message" : message})
	if err!= nil {
		log.Fatalf("Error storing message in Mongo  %v", err)
	}
	return err
} 

func GetChatHistoryMessageFromMongo(roomID string , limit int64) ([]string , error){
	mongoClient = CreateMongoClient()
	defer mongoClient.Disconnect(ctx)
	collection := mongoClient.Database("Chat-DB").Collection("messages")
	history , err := collection.Find(ctx , bson.M{"roomID":roomID}, &options.FindOptions{Sort: bson.M{"_id": -1},Limit: &limit})
	if err!= nil {
		return nil , err
	}
	defer history.Close(ctx)
	var messages []string
	for history.Next(ctx) {
		var result struct{
			Message string `bson:"message"`
		}
		if err := history.Decode(&result);err!=nil{
			return nil ,err
		}
		messages = append(messages, result.Message)
	}
	return messages , nil

}