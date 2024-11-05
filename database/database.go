package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    if err != nil {
        log.Fatal(err)
    } 
	// Set up the database connection
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err!=nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

var Client *mongo.Client = DBSet()

func CollectionDB(collectionName string) *mongo.Collection {
	// Set up the database connection
	collection := Client.Database(os.Getenv("DB_NAME")).Collection(collectionName)
	return collection
}