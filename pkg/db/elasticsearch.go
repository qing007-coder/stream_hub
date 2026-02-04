package db

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"stream_hub/pkg/model/config"
)

type ElasticSearchClient struct {
	client *elasticsearch.TypedClient
}

func NewElasticSearchClient(conf *config.CommonConfig) (*ElasticSearchClient, error) {
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("%s:%s", conf.Elasticsearch.Addr, conf.Elasticsearch.Port)},
	})

	if err != nil {
		return nil, err
	}

	return &ElasticSearchClient{client: client}, nil
}

func (e *ElasticSearchClient) Client() *elasticsearch.TypedClient {
	return e.client
}
