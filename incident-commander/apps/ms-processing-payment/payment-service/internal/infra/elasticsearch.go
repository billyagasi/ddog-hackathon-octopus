package infra

import (
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/config"
)

func NewElasticsearchClient(cfg config.Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
		Transport: httptrace.WrapRoundTripper(http.DefaultTransport),
	}

	if cfg.ElasticsearchUser != "" && cfg.ElasticsearchPassword != "" {
		esCfg.Username = cfg.ElasticsearchUser
		esCfg.Password = cfg.ElasticsearchPassword
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	res, err := es.Ping()
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	log.Println("[infra] Elasticsearch connected")
	return es, nil
}
