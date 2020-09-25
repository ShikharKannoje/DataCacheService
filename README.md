"# dataCacheService" 

Two services written, APIGatway and Cache service.

1) All the API are exposed in APIGateway,
2) All the APIs to caching serice reaches using APIGateway.
3) API Gatway basically make network calls to caching service as the both are different microservices.
4) Swagger for APIGateway has been uploaded.


The cache policy is LRU (Least Recently Used)


Performance of the APIs

Time taken for response
    Cache Hit: 25ms
    Loss:   450ms


