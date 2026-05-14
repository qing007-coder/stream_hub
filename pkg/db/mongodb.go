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

	uri := fmt.Sprintf("mongodb://%s:%s", conf.MongoDB.Addr, conf.MongoDB.Port)

	if conf.MongoDB.Username != "" && conf.MongoDB.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s",
			conf.MongoDB.Username,
			conf.MongoDB.Password,
			conf.MongoDB.Addr,
			conf.MongoDB.Port,
		)
	}

	clientOptions := options.Client().ApplyURI(uri)

	if conf.MongoDB.Database != "" {
		clientOptions.SetAuth(options.Credential{
			AuthSource: conf.MongoDB.Database,
			Username:   conf.MongoDB.Username,
			Password:   conf.MongoDB.Password,
		})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	return &MongoClient{client: client}, nil
}

func (m *MongoClient) Client() *mongo.Client {
	return m.client
}
