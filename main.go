package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	coingecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var CG = coingecko.NewClient(httpClient)

func fetchForCoin(coinID string) {
	coin, err := CG.CoinsID(coinID, true, true, true, true, true, true)
	log.Debugf("update %s %s", coinID, rConf.currency)
	if err != nil {
		log.Fatal(err)
	}

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

	var baseURL = "https://api.coingecko.com/api/v3/coins"
	resp, _ :=CG.MakeReq(baseURL)


	//@todo PR into the library
	var data *types.CoinList
	err :=json.Unmarshal(resp, &data)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(rConf.updateInterval)
	for {
		log.Debug("> Updating....\n")
		for _, item := range *data {
		//for _, item := range []string{"bitcoin"} {

			fetchForCoin(item.ID)
			time.Sleep(time.Second *2 ) //@todo better api handling of api throttling
		}
		<-ticker.C
	}
}
