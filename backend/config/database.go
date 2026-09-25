package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB holds references to MongoDB collections
type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Users    *mongo.Collection
	Polls    *mongo.Collection
	Votes    *mongo.Collection
}

// ConnectDatabase initializes the MongoDB connection using the configured URI
func ConnectDatabase() (*DB, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "pulsevote"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("[MongoDB] Successfully connected to database: %s", dbName)

	database := client.Database(dbName)
	usersCol := database.Collection("users")
	pollsCol := database.Collection("polls")
	votesCol := database.Collection("votes")

	// Ensure unique index on user email
	_, err = usersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("[MongoDB] Notice: user email index setup: %v", err)
	}

	// Ensure unique index on poll shareCode
	_, err = pollsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "shareCode", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("[MongoDB] Notice: poll shareCode index setup: %v", err)
	}

	return &DB{
		Client:   client,
		Database: database,
		Users:    usersCol,
		Polls:    pollsCol,
		Votes:    votesCol,
	}, nil
}
