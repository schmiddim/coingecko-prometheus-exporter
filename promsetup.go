package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type runtimeConfStruct struct {
	registry                  *prometheus.Registry
	vectors                   []*prometheus.GaugeVec
	httpServerPort            uint
	httpServ                  *http.Server
	updateInterval            time.Duration
	debug                     bool
	configFile                string
	currency                  string
	currentPrice              *prometheus.GaugeVec
	ath                       *prometheus.GaugeVec
	athRelative               *prometheus.GaugeVec
	change24h                 *prometheus.GaugeVec
	change7d                  *prometheus.GaugeVec
	change14d                 *prometheus.GaugeVec
	change30d                 *prometheus.GaugeVec
	change60d                 *prometheus.GaugeVec
	change200d                *prometheus.GaugeVec
	marketCap                 *prometheus.GaugeVec
	marketCapChange24Relative *prometheus.GaugeVec
	high24                    *prometheus.GaugeVec
	low24                     *prometheus.GaugeVec
}

var rConf = runtimeConfStruct{

	httpServerPort:            9101,
	httpServ:                  nil,
	registry:                  prometheus.NewRegistry(),
	updateInterval:            5 * time.Second,
	configFile:                "",
	currency:                  "eur",
	currentPrice:              nil,
	ath:                       nil,
	athRelative:               nil,
	change24h:                 nil,
	change7d:                  nil,
	change14d:                 nil,
	change30d:                 nil,
	change60d:                 nil,
	change200d:                nil,
	marketCap:                 nil,
	marketCapChange24Relative: nil,
	high24:                    nil,
	low24:                     nil,
}

func setupGauges() {

	// Init Prometheus Gauge Vectors
	rConf.currentPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "current_price",
		Help:      fmt.Sprintf("current price")}, []string{"symbol"})
	rConf.ath = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "ath",
		Help:      fmt.Sprintf("alltime high")}, []string{"symbol"})
	rConf.athRelative = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "ath_relative",
		Help:      fmt.Sprintf("alltime high relative")}, []string{"symbol"})
	rConf.change24h = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change24h",
		Help:      fmt.Sprintf("change24h")}, []string{"symbol"})
	rConf.change7d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change7d",
		Help:      fmt.Sprintf("change7d")}, []string{"symbol"})
	rConf.change14d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change14d",
		Help:      fmt.Sprintf("change14d")}, []string{"symbol"})
	rConf.change30d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change30d",
		Help:      fmt.Sprintf("change30d")}, []string{"symbol"})

	rConf.change60d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change60d",
		Help:      fmt.Sprintf("change60d")}, []string{"symbol"})
	rConf.change200d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change200d",
		Help:      fmt.Sprintf("change200d")}, []string{"symbol"})
	rConf.marketCap = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "marketCap",
		Help:      fmt.Sprintf("marketCap")}, []string{"symbol"})
	rConf.marketCapChange24Relative = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "marketCapChange24Relative",
		Help:      fmt.Sprintf("marketCapChange24Relative")}, []string{"symbol"})
	rConf.high24 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "high24",
		Help:      fmt.Sprintf("high24")}, []string{"symbol"})

	rConf.low24 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "low24",
		Help:      fmt.Sprintf("low24")}, []string{"symbol"})
	rConf.registry.MustRegister(rConf.currentPrice)
	rConf.registry.MustRegister(rConf.ath)
	rConf.registry.MustRegister(rConf.athRelative)
	rConf.registry.MustRegister(rConf.change24h)
	rConf.registry.MustRegister(rConf.change7d)
	rConf.registry.MustRegister(rConf.change14d)
	rConf.registry.MustRegister(rConf.change30d)
	rConf.registry.MustRegister(rConf.change60d)
	rConf.registry.MustRegister(rConf.change200d)
	rConf.registry.MustRegister(rConf.marketCap)
	rConf.registry.MustRegister(rConf.marketCapChange24Relative)
	rConf.registry.MustRegister(rConf.high24)
	rConf.registry.MustRegister(rConf.low24)
}

func initParams() {

	flag.UintVar(&rConf.httpServerPort, "httpServerPort", rConf.httpServerPort, "HTTP server port.")
	flag.BoolVar(&rConf.debug, "debug", false, "Set debug log level.")
	flag.StringVar(&rConf.configFile, "currency", "eur", "currency")

	flag.Parse()

	logLvl := log.InfoLevel
	if rConf.debug {
		logLvl = log.DebugLevel
	}
	log.SetLevel(logLvl)

}

func setupWebserver() {
	// Register prom metrics path in http serv
	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		rConf.registry,
		promhttp.HandlerFor(rConf.registry, promhttp.HandlerOpts{}),
	))

	// Init & start serv
	rConf.httpServ = &http.Server{
		Addr:    fmt.Sprintf(":%d", rConf.httpServerPort),
		Handler: httpMux,
	}
	go func() {
		log.Infof("> Starting HTTP server at %s\n", rConf.httpServ.Addr)
		err := rConf.httpServ.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Errorf("HTTP Server errored out %v", err)
		}
	}()

}
