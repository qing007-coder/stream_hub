package infra

import (
	"go.mongodb.org/mongo-driver/mongo"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
)

type Mongo struct {
	client *mongo.Client
}

func NewMongo(conf *config.CommonConfig) (*Mongo, error) {
	m := new(Mongo)
	client, err := db.NewMongoClient(conf)
	if err != nil {
		return nil, err
	}

	m.client = client.Client()
	return m, nil
}

func (m *Mongo) Collection(database, collection string) *mongo.Collection {
	return m.client.Database(database).Collection(collection)
}
