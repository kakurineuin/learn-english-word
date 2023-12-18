package database

import (
	"context"
	"fmt"
	"time"

	"github.com/kakurineuin/learn-english-word/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func ConnectDB() error {
	newClient, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(config.EnvMongoDBURI()).SetTimeout(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("ConnectDB failed! error: %w", err)
	}

	client = newClient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("ConnectDB ping database failed! error: %w", err)
	}

	fmt.Println("Connected to MongoDB")
	return nil
}

func DisconnectDB() error {
	if err := client.Disconnect(context.TODO()); err != nil {
		return fmt.Errorf("DisconnectDB failed! error: %w", err)
	}

	return nil
}

func GetCollection(collectionName string) *mongo.Collection {
	collection := client.Database("learnEnglish").Collection(collectionName)
	return collection
}
