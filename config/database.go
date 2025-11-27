package config
import (
	"context"
	"fmt"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Ganti dengan connection string MongoDB Anda
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Test connection
err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

fmt.Println("Connected to MongoDB!")

// Set database
DB = client.Database("upload_db")
}