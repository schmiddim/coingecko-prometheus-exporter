package main

import (
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
	coingecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type runtimeConfStruct struct {
	debug                bool
	sleepAfterRequest    int
	sleepAfterThrottling int
	currency             string
}

var rConf = runtimeConfStruct{
	sleepAfterThrottling: 15000,
	sleepAfterRequest:    2500,
	currency:             "eur",
}

var CG = coingecko.NewClient(httpClient)

func fetchForCoin(coinID string) {
	coin, err := CG.CoinsID(coinID, true, true, true, false, false, true)
	log.Debugf("update %s %s", coinID, rConf.currency)
	if err != nil || coin == nil {
		log.Errorf("Loop: We're throttled by API, %s", err)
		time.Sleep(time.Millisecond * time.Duration(rConf.sleepAfterThrottling))
		fetchForCoin(coinID)

	}
	if coin == nil {
		log.Errorf("WTF: %s Coin is nil, %s", coinID, coin)
		time.Sleep(time.Second * 15)
		fetchForCoin(coinID)

	}

	prometheusConfig.currentPrice.WithLabelValues(coin.Symbol).Set(coin.MarketData.CurrentPrice[rConf.currency])
	prometheusConfig.ath.WithLabelValues(coin.Symbol).Set(coin.MarketData.ATH[rConf.currency])
	prometheusConfig.athRelative.WithLabelValues(coin.Symbol).Set(coin.MarketData.ATHChangePercentage[rConf.currency])
	prometheusConfig.change24h.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChange24h)
	prometheusConfig.change7d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage7d)
	prometheusConfig.change14d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage14d)
	prometheusConfig.change30d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage30d)
	prometheusConfig.change60d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage60d)
	prometheusConfig.change200d.WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage200d)
	prometheusConfig.marketCapChange24Relative.WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCapChange24h)
	prometheusConfig.marketCap.WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCap[rConf.currency])
	prometheusConfig.high24.WithLabelValues(coin.Symbol).Set(coin.MarketData.High24[rConf.currency])
	prometheusConfig.low24.WithLabelValues(coin.Symbol).Set(coin.MarketData.Low24[rConf.currency])
	time.Sleep(time.Duration(rConf.sleepAfterRequest) * time.Millisecond) //@todo better api handling of api throttling
}
func initParams() {

	flag.UintVar(&prometheusConfig.httpServerPort, "httpServerPort", prometheusConfig.httpServerPort, "HTTP server port.")
	flag.BoolVar(&rConf.debug, "debug", false, "Set debug log level.")
	flag.StringVar(&rConf.currency, "currency", "eur", "currency")
	flag.IntVar(&rConf.sleepAfterRequest, "sleepAfterRequest", rConf.sleepAfterRequest, "Time in ms to wait after each coin request")
	flag.IntVar(&rConf.sleepAfterThrottling, "sleepAfterRequest", rConf.sleepAfterThrottling, "Time in ms to wait after each coin request")

	flag.Parse()

	logLvl := log.InfoLevel
	if rConf.debug {
		logLvl = log.DebugLevel
	}
	log.SetLevel(logLvl)

}
func main() {

	initParams()
	setupWebserver()
	setupGauges()

	exec()

}
func exec() {
	var baseURL = "https://api.coingecko.com/api/v3/coins"
	resp, err := CG.MakeReq(baseURL)

	if err != nil {
		sleepInterval := time.Millisecond * time.Duration(rConf.sleepAfterThrottling)
		log.Errorf("Init: We're throttled by API, %s  - wait %d", err, sleepInterval)
		time.Sleep(sleepInterval)

	}
	//@todo PR into the library
	var data *types.CoinList
	err = json.Unmarshal(resp, &data)
	if err != nil {
		log.Fatal(err)
		exec()
	}

	for {
		log.Debug("> Updating....")
		for _, item := range *data {
			fetchForCoin(item.ID)
		}
	}
}
