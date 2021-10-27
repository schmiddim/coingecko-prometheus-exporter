# Prometheus CoinGecko Exporter

This exporter fetches the top 50 coins from the [CoinGecko API](https://www.coingecko.com/de/api) and exports some
interesting metrics.

## Usage

```
go get 
go build
./coingecko-exporter 
curl http://localhost:9101
```


## Issues
- No support for paid plans so far
- You can run in throttling trouble (50 calls / Minute)
