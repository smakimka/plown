package elastic

import "github.com/elastic/go-elasticsearch/v9"

type Config struct {
	Addr string
}

func New(cfg Config) (*elasticsearch.TypedClient, error) {
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{cfg.Addr},
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}
