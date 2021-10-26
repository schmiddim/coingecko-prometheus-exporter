package main

import (
	log "github.com/sirupsen/logrus"
	coingecko "github.com/superoo7/go-gecko/v3"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var CG = coingecko.NewClient(httpClient)

func fetchForCoin(coinID string) {
	coin, err := CG.CoinsID(coinID, true, true, true, true, true, true)
	log.Debug(coinID, coin.MarketData.CurrentPrice["eur"])
	if err != nil {
		log.Fatal(err)
	}

	//cap24 market cap

	rConf.currentPrice.WithLabelValues(coin.Symbol).Set(coin.MarketData.CurrentPrice[rConf.currency])
	rConf.ath.WithLabelValues(coin.Symbol).Set(coin.MarketData.ATH[rConf.currency])
	rConf.athRelative.WithLabelValues(coin.Symbol).Set(coin.MarketData.ATHChangePercentage[rConf.currency])
	rConf.change24h.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChange24h)
	rConf.change7d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage7d)
	rConf.change14d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage14d)
	rConf.change30d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage30d)
	rConf.change60d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage60d)
	rConf.change200d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage200d)
	rConf.marketCapChange24Relative.WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCapChange24h)
	rConf.marketCap.WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCap[rConf.currency])
	rConf.high24.WithLabelValues(coin.Symbol).Set(coin.MarketData.High24[rConf.currency])
	rConf.low24.WithLabelValues(coin.Symbol).Set(coin.MarketData.Low24[rConf.currency])

}

func main() {

	initParams()
	setupWebserver()
	setupGauges()

	// Regular loop operations below
	ticker := time.NewTicker(rConf.updateInterval)
	for {
		log.Debug("> Updating....\n")
		for _, item := range []string{"bitcoin"} {
			fetchForCoin(item)
		}
		<-ticker.C
	}
}
