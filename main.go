package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracingLog "github.com/opentracing/opentracing-go/log"
	log "github.com/sirupsen/logrus"
	coinGecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"golang.org/x/time/rate"
	"net/http"
	"sort"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type runtimeConfStruct struct {
	debug                bool
	requestsPerMinute    int
	sleepAfterThrottling int
	currency             string
	additionalCoins      []string
}

var rConf = runtimeConfStruct{
	debug:                true,
	sleepAfterThrottling: 15000,
	currency:             "eur",
	requestsPerMinute:    49,
}

var CG = coinGecko.NewClient(httpClient)

func main() {

	ctx := initTracing()
	initParams(ctx)
	setupWebserver(ctx)
	setupGauges(ctx)

	fmt.Printf("debug mode %t\n", rConf.debug)
	fmt.Printf("currency %s\n", rConf.currency)
	fmt.Printf("requestsPerMinute %d\n", rConf.requestsPerMinute)
	fmt.Printf("sleepAfterThrottling %d\n", rConf.sleepAfterThrottling)
	exec(ctx)

}
func contains(s []string, searchTerm string) bool {
	i := sort.SearchStrings(s, searchTerm)
	return i < len(s) && s[i] == searchTerm
}

func initParams(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "initParams")
	defer span.Finish()
	additionalCoinsString := "foo"

	flag.UintVar(&prometheusConfig.httpServerPort, "httpServerPort", prometheusConfig.httpServerPort, "HTTP server port.")
	flag.BoolVar(&rConf.debug, "debug", rConf.debug, "Set debug log level.")
	flag.StringVar(&rConf.currency, "currency", "eur", "currency")
	flag.IntVar(&rConf.requestsPerMinute, "requestsPerMinute", rConf.requestsPerMinute, "how many requestsPerMinute ")
	flag.IntVar(&rConf.sleepAfterThrottling, "sleepAfterRequest", rConf.sleepAfterThrottling, "Time in ms to wait after each coin request")
	flag.StringVar(&additionalCoinsString, "additionalCoins", "", "pass additional coins coma separated")
	flag.Parse()

	rConf.additionalCoins = strings.Split(additionalCoinsString, ",")
	logLvl := log.InfoLevel
	if rConf.debug {
		logLvl = log.DebugLevel
	}
	log.SetLevel(logLvl)

}

func exec(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "exec")
	defer span.Finish()

	var baseURL = "https://api.coingecko.com/api/v3/coins"
	resp, err := CG.MakeReq(baseURL)

	if err != nil {
		sleepInterval := time.Millisecond * time.Duration(rConf.sleepAfterThrottling)
		log.Errorf("Init: We're throttled by API, %s  - wait %d", err, sleepInterval)
		ext.LogError(span, err)
		time.Sleep(sleepInterval)
		exec(ctx)
		return

	}
	//@todo PR into the library
	var data *types.CoinList
	err = json.Unmarshal(resp, &data)
	if err != nil {
		ext.LogError(span, err)
		log.Error(err)
		exec(ctx)
		return
	}

	var symbols []string
	log.Debug("Updating....")
	for _, item := range *data {
		symbols = append(symbols, item.ID)
	}
	sort.Strings(symbols)
	for _, item := range rConf.additionalCoins {
		if !contains(symbols, strings.TrimSpace(item)) {
			symbols = append(symbols, strings.TrimSpace(item))
		}
	}

	n := rate.Every(time.Minute / time.Duration(rConf.requestsPerMinute))
	limiter := rate.NewLimiter(n, 1)
	ctxBackground := context.Background()
	for {
		log.Debug("> Updating....")
		for _, item := range symbols {
			if err := limiter.Wait(ctxBackground); err != nil {
				log.Fatalln(err)
			}
			fetchForCoin(ctx, item)

		}
	}
}
func fetchForCoin(ctx context.Context, coinID string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "fetchForCoin")
	defer span.Finish()
	span.SetTag("coinID", coinID)
	span.LogFields(
		tracingLog.String("coinID", coinID),
	)

	coin, err := CG.CoinsID(coinID, true, true, true, false, false, true)
	log.Debugf("update %s %s", coinID, rConf.currency)
	if err != nil || coin == nil {
		if err.Error() == "{\"error\":\"Could not find coin with the given id\"}" {
			log.Fatalf("coinID '%s' not found - please pass only Coingecko IDs and not Symbols", coinID)
		}
		ext.LogError(span, err)
		log.Errorf("Loop: We're throttled by API, %s", err)
		time.Sleep(time.Millisecond * time.Duration(rConf.sleepAfterThrottling))
		coin, err = nil, nil
		fetchForCoin(ctx, coinID)
		return
	}

	prometheusConfig.gaugeVectors["currentPrice"].WithLabelValues(coin.Symbol).Set(coin.MarketData.CurrentPrice[rConf.currency])
	prometheusConfig.gaugeVectors["ath"].WithLabelValues(coin.Symbol).Set(coin.MarketData.ATH[rConf.currency])
	prometheusConfig.gaugeVectors["athRelative"].WithLabelValues(coin.Symbol).Set(coin.MarketData.ATHChangePercentage[rConf.currency])
	prometheusConfig.gaugeVectors["change1h"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage1hInCurrency[rConf.currency])
	prometheusConfig.gaugeVectors["change24h"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage24h)
	prometheusConfig.gaugeVectors["change7d"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage7d)
	prometheusConfig.gaugeVectors["change14d"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage14d)
	prometheusConfig.gaugeVectors["change30d"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage30d)
	prometheusConfig.gaugeVectors["change60d"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage60d)
	prometheusConfig.gaugeVectors["change200d"].WithLabelValues(coin.Symbol).Set(coin.MarketData.PriceChangePercentage200d)
	prometheusConfig.gaugeVectors["marketCapChange24Relative"].WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCapChangePercentage24h)
	prometheusConfig.gaugeVectors["marketCap"].WithLabelValues(coin.Symbol).Set(coin.MarketData.MarketCap[rConf.currency])
	prometheusConfig.gaugeVectors["high24"].WithLabelValues(coin.Symbol).Set(coin.MarketData.High24[rConf.currency])
	prometheusConfig.gaugeVectors["low24"].WithLabelValues(coin.Symbol).Set(coin.MarketData.Low24[rConf.currency])

}
