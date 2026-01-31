package infra

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
)

type Elasticsearch struct {
	ctx    context.Context
	client *elasticsearch.TypedClient
	index  string
}

func NewElasticSearch(conf *config.CommonConfig, index string) (*Elasticsearch, error) {
	es := new(Elasticsearch)
	es.ctx = context.Background()
	es.index = index
	client, err := db.NewElasticSearchClient(conf)
	if err != nil {
		return nil, err
	}

	es.client = client.Client()
	return es, nil
}

func (es *Elasticsearch) Search(mustQueries []types.Query, shouldQueries []types.Query, sort []types.SortCombinations, from, size int) (*search.Response, error) {
	return es.client.Search().
		Index(es.index).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Must:   mustQueries,   // 每个都要匹配
					Should: shouldQueries, // 有一个匹配就可以
				},
			},
			Sort: sort,
			Size: &size,
			From: &from,
		}).Do(es.ctx)
}

func (es *Elasticsearch) Update(id string, query map[string]interface{}) error {
	data, _ := json.Marshal(&query)
	_, err := es.client.Update(es.index, id).
		Request(&update.Request{
			Doc: data,
		}).Do(context.TODO())

	return err
}
