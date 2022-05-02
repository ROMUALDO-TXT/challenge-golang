package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client
var database *mongo.Database
var mongoCtx context.Context

func GetCollection(collectionName string) *mongo.Collection {
	return database.Collection(collectionName)
}

func GetContext() context.Context {
	return mongoCtx
}

func CloseConnection() {
	fmt.Println("Closing MongoDB connection")
}

func CreateConnection() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URL")))
	if err != nil {
		log.Fatalf("Couldn't create MongoDB Client: %v", err)
	}

	mongoCtx = context.Background()

	if err = client.Connect(mongoCtx); err != nil {
		log.Fatalf("Couldn't connect to MongoDB: %v", err)
	}

	if err = client.Ping(mongoCtx, readpref.Primary()); err != nil {
		log.Fatalf("Couldn't reach MongoDB: %v", err)
	}

	database = client.Database("Klever-blog")

	fmt.Println("Connected to MongoDB!")
}
