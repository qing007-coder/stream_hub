package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"stream_hub/pkg/model/config"
)

type MongoClient struct {
	client *mongo.Client
}

func NewMongoClient(conf *config.CommonConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s", conf.MongoDB.Addr, conf.MongoDB.Port)))
	if err != nil {
		return nil, err
	}

	return &MongoClient{client: client}, nil
}

func (m *MongoClient) Client() *mongo.Client {
	return m.client
}
