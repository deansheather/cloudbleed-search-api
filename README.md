# Cloudbleed Search API
Search pirate/sites-using-cloudflare for a domain name using Elasticsearch,
exposed as a queryable API.

Public instance: https://cloudbleed-api.whats-th.is (running under CloudFlare
:^))

### Usage
1. Launch Elasticsearch and make the `domains` index (see below)
2. Add your data  (see `import-data.go`)
3. Build `cloudbleed.go`
4. Run (use environment variables `PORT` and `ELASTIC_ENDPOINT` to change
   default config variables)
5. Enjoy

### `domains` index
```json
{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "cloudbleed_domain":{
            "properties":{
                "domain":{
                    "type": "text",
                    "index": true,
                    "store": true,
                    "fielddata": true
                }
            }
        }
    }
}
```

### License
A copy of the MIT license can be found in `LICENSE`.
