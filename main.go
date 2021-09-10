package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Categories  string `json:"categories"`
	ClientId    string `json:"client_id"`
}

type SearchResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchHit
	}
}

type SearchHit struct {
	Score   float64 `json:"_score"`
	Index   string  `json:"_index"`
	Type    string  `json:"_type"`
	Version int64   `json:"_version,omitempty"`

	Source Item `json:"_source"`
}

const (
	indexName = "items"
)

func main() {
	es, err := elasticsearch7.NewClient(
		elasticsearch7.Config{
			Addresses: []string{"http://127.0.0.1:9200"},
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	res, err := es.Ping()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	// Check response status
	if res.IsError() || res.StatusCode != http.StatusOK {
		log.Fatalf("Error getting response: %s", res.String())

	}
	fmt.Println(res.Body)

	indexMapping := `
{
	"settings": {
		"analysis": {
			"analyzer": {
				"custom_edge_ngram_analyzer": {
					"type": "custom",
					"tokenizer": "customized_edge_tokenizer",
					"filter": [
						"lowercase"
					]
				}
			},
			"tokenizer": {
				"customized_edge_tokenizer": {
					"type": "edge_ngram",
					"min_gram": 3,
					"max_gram": 10,
					"token_chars": [
						"letter",
						"digit"
					]
				}
			}
		}
	},
	"mappings": {
		"dynamic": true,
		"properties": {
			"id": {
				"type": "text",
				"analyzer": "custom_edge_ngram_analyzer"
			},
			"name": {
				"type": "text",
				"analyzer": "custom_edge_ngram_analyzer"
			},
			"description": {
				"type": "text"
			},
			"categories": {
				"type": "text"
			},
			"client_id": {
				"type": "text"
			}
		}
	}
}`
	res, err = es.Indices.Create(indexName,
		es.Indices.Create.WithBody(strings.NewReader(indexMapping)),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	// Check response status
	if res.IsError() || res.StatusCode != http.StatusOK {
		log.Fatalf("Error getting response: %s", res.String())

	}
	fmt.Println(res.StatusCode)
	result, err := ListItem(es, "client-A", "item1")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println(result)

}

func ListItem(es *elasticsearch7.Client, cliendId, query string) ([]Item, error) {
	queryReq := fmt.Sprintf(`
	{
		"query": {
			"bool": {
				"must": [{
					"match": {
						"tenant_id": "%s"
					}
				}, {
					"multi_match": {
						"query": "%s",
						"type": "best_fields",
						"fields": [
							"name",
							"id"
						],
						"operator":   "or"
					}
				}]
			}
		}
	}`, cliendId, query)
	res, err := es.Search(
		es.Search.WithIndex(indexName),
		es.Search.WithBody(strings.NewReader(queryReq)),
		es.Search.WithPretty(),
	)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	// Check response status
	if res.IsError() || res.StatusCode != http.StatusOK {
		return nil, err

	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var sr SearchResponse
	if err := json.Unmarshal([]byte(body), &sr); err != nil {
		return nil, err

	}
	entities := make([]Item, 0)

	for _, hit := range sr.Hits.Hits {
		entities = append(entities, hit.Source)
	}

	return entities, nil
}
