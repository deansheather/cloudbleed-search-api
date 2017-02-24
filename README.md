# Cloudbleed Search API
Search pirate/sites-using-cloudflare for a domain name using Elasticsearch,
exposed as a queryable API.

Public instance: https://cloudbleed-api.whats-th.is (running under CloudFlare
:^) - this is also ratelimited)

### Usage
1. Launch Elasticsearch and make the `domains` index (see below)
2. Import your data (see below)
3. Build `cloudbleed-search-api.go`
4. Run (use environment variables `PORT` and `ELASTIC_ENDPOINT` to change
   default config variables)
5. Enjoy

### `domains` index
```json
{
    "mappings": {
        "cloudbleed_domain":{
            "properties":{
                "domain":{
                    "type": "text",
                    "analyzer": "keyword",
                    "fielddata": true,
                    "index": true,
                    "index_options": "docs",
                    "store": true
                }
            }
        }
    }
}
```

### Importing data
1. Download the latest text file of CloudFlare domains as `/import.txt`
2. Run `tail -c +2 /import.txt > /temp.txt && mv /temp.txt /import.txt` (remove
   leading empty line)
3. Run `logstash -f logstash-import` (after changing Elasticsearch host)
4. Wait a while (until Elastic CPU drops to idle again)

### License
A copy of the MIT license can be found in `LICENSE`.
