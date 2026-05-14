package db

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"stream_hub/pkg/model/config"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticSearchClient struct {
	client *elasticsearch.TypedClient
}

func NewElasticSearchClient(conf *config.CommonConfig) (*ElasticSearchClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("%s:%s", conf.Elasticsearch.Addr, conf.Elasticsearch.Port)},
	}

	if conf.Elasticsearch.Username != "" && conf.Elasticsearch.Password != "" {
		esCfg.Username = conf.Elasticsearch.Username
		esCfg.Password = conf.Elasticsearch.Password
		esCfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client, err := elasticsearch.NewTypedClient(esCfg)
	if err != nil {
		return nil, err
	}

	return &ElasticSearchClient{client: client}, nil
}

func (e *ElasticSearchClient) Client() *elasticsearch.TypedClient {
	return e.client
}
