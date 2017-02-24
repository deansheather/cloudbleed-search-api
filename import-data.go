package main

import (
	"bufio"
	"context"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	elastic "gopkg.in/olivere/elastic.v5"
)

type cloudbleedDomain struct {
	Domain string `json:"domain"`
}

func main() {
	// Create Elastic client context
	ectx := context.Background()

	// Create Elastic client
	elasticEndpoint := os.Getenv("ELASTIC_ENDPOINT")
	if elasticEndpoint == "" {
		elasticEndpoint = "http://127.0.0.1:9200"
	}
	el, err := elastic.NewClient(elastic.SetURL(elasticEndpoint))
	if err != nil {
		log.WithField("error", err).Fatal("failed to create Elastic client")
		return
	}
	info, code, err := el.Ping(elasticEndpoint).Do(ectx)
	if err != nil {
		log.WithField("error", err).Fatal("failed to ping Elasticsearch server")
		return
	}
	log.Infof("pinged Elasticsearch %s with code %d", info.Version.Number, code)

	// Parse file
	if file, err := os.Open("import.txt"); err == nil {
		defer file.Close()

		// Scan file
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			_, err := el.Index().
				Index("domains").
				Type("cloudbleed_domain").
				Id(uuid.NewV4().String()).
				BodyJson(cloudbleedDomain{Domain: scanner.Text()}).
				Do(ectx)
			if err != nil {
				log.WithField("error", err).Errorf("failed to add item \"%v\" to the Elasticsearch index")
			}
		}

		// Check for errors
		if err = scanner.Err(); err != nil {
			log.WithField("error", err).Fatal("failed to scan file import.txt")
			return
		}

		log.Info("All done!")
	} else {
		log.WithField("error", err).Fatal("failed to open file import.txt")
	}
}
