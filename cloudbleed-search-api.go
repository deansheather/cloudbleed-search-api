package main

import (
	"context"
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/cors"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Max search string length
const maxQueryLength = 30

// Search response struct returned on search API call.
type searchResponse struct {
	Query   string   `json:"query"`
	Results int64    `json:"results"`
	Items   []string `json:"items"`
}

// Raw data source field.
type sourceField struct {
	Domain string `json:"domain"`
}

// Error struct returned when the API errors.
type apiError struct {
	Error bool   `json:"error"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}

// API errors.
var (
	invalidSearchQueryError = apiError{Error: true, Code: iris.StatusBadRequest, Msg: "Bad Request: invalid search query"}
	failedToGetResultError  = apiError{Error: true, Code: iris.StatusInternalServerError, Msg: "Internal Server Error: failed to get result from Elasticsearch"}
)

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
	log.Debugf("pinged Elasticsearch %s with code %d", info.Version.Number, code)

	// Create Iris app
	app := iris.New()

	// Configure Iris app
	app.Adapt(httprouter.New())                                      // Router
	app.Adapt(view.Handlebars("./views", ".handlebars"))             // View engine
	app.Adapt(cors.New(cors.Options{AllowedOrigins: []string{"*"}})) // CORS

	// > GET /api/v1/search
	// Search for the domain in the Cloudbleed list. Use `?query` query
	// parameter to denote search term. Must be a domain.
	app.Get("/api/v1/search", func(ctx *iris.Context) {
		q := ctx.URLParam("query")
		if q == "" || len(q) > maxQueryLength {
			ctx.JSON(iris.StatusBadRequest, invalidSearchQueryError)
			return
		}

		res, err := el.Search().
			Index("domains").
			Query(elastic.NewTermQuery("domain", q)).
			Sort("domain", true).
			From(0).Size(10).
			Do(ectx)
		if err != nil {
			log.WithField("error", err).Error("failed to get result from Elasticsearch")
			ctx.JSON(iris.StatusInternalServerError, failedToGetResultError)
			return
		}
		log.Debugf("query took %vms", res.TookInMillis)
		if res.TotalHits() > 0 {
			output := []string{}
			for _, hit := range res.Hits.Hits {
				var d sourceField
				err := json.Unmarshal(*hit.Source, &d)
				if err == nil {
					output = append(output, d.Domain)
				}
			}
			ctx.JSON(200, searchResponse{
				Query:   q,
				Results: res.TotalHits(),
				Items:   output,
			})
		} else {
			ctx.JSON(200, searchResponse{
				Query:   q,
				Results: int64(0),
				Items:   []string{},
			})
		}
	})

	// 404 handler
	app.OnError(404, func(ctx *iris.Context) {
		ctx.JSON(404, map[string]interface{}{
			"code": 404,
			"msg":  "Search API can be found at /api/v1/search?query=domain.tld",
		})
	})

	// Start the server on port in env (or 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Infof("trying to listen for requests on port %v", port)
	app.Listen(":" + port)
}
