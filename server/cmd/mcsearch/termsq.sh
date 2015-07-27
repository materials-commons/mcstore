#!/bin/sh

curl -XGET http://localhost:9200/mc/files/_search -d '
{
    "query": { 
         "bool": { 
             "must": [ 
                    { "term": {"project_id": "d232df78-cbe2-4561-a958-7fd45b87601d"} },
                        { "query_string": {"query": "contents:(sigma OR ion)"} }
              ]
          }
      }
}' | jq .
