package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type prometheusConfigStruct struct {
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

var prometheusConfig = prometheusConfigStruct{
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
	registry:                  prometheus.NewRegistry(),
}

func setupGauges() {

	// Init Prometheus Gauge Vectors
	prometheusConfig.currentPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "current_price",
		Help:      fmt.Sprintf("current price")}, []string{"symbol"})
	prometheusConfig.ath = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "ath",
		Help:      fmt.Sprintf("alltime high")}, []string{"symbol"})
	prometheusConfig.athRelative = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "ath_relative",
		Help:      fmt.Sprintf("alltime high relative")}, []string{"symbol"})
	prometheusConfig.change24h = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change24h",
		Help:      fmt.Sprintf("change24h")}, []string{"symbol"})
	prometheusConfig.change7d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change7d",
		Help:      fmt.Sprintf("change7d")}, []string{"symbol"})
	prometheusConfig.change14d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change14d",
		Help:      fmt.Sprintf("change14d")}, []string{"symbol"})
	prometheusConfig.change30d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change30d",
		Help:      fmt.Sprintf("change30d")}, []string{"symbol"})

	prometheusConfig.change60d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change60d",
		Help:      fmt.Sprintf("change60d")}, []string{"symbol"})
	prometheusConfig.change200d = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "change200d",
		Help:      fmt.Sprintf("change200d")}, []string{"symbol"})
	prometheusConfig.marketCap = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "marketCap",
		Help:      fmt.Sprintf("marketCap")}, []string{"symbol"})
	prometheusConfig.marketCapChange24Relative = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "marketCapChange24Relative",
		Help:      fmt.Sprintf("marketCapChange24Relative")}, []string{"symbol"})
	prometheusConfig.high24 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "high24",
		Help:      fmt.Sprintf("high24")}, []string{"symbol"})

	prometheusConfig.low24 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "coin_gecko",
		Name:      "low24",
		Help:      fmt.Sprintf("low24")}, []string{"symbol"})
	prometheusConfig.registry.MustRegister(prometheusConfig.currentPrice)
	prometheusConfig.registry.MustRegister(prometheusConfig.ath)
	prometheusConfig.registry.MustRegister(prometheusConfig.athRelative)
	prometheusConfig.registry.MustRegister(prometheusConfig.change24h)
	prometheusConfig.registry.MustRegister(prometheusConfig.change7d)
	prometheusConfig.registry.MustRegister(prometheusConfig.change14d)
	prometheusConfig.registry.MustRegister(prometheusConfig.change30d)
	prometheusConfig.registry.MustRegister(prometheusConfig.change60d)
	prometheusConfig.registry.MustRegister(prometheusConfig.change200d)
	prometheusConfig.registry.MustRegister(prometheusConfig.marketCap)
	prometheusConfig.registry.MustRegister(prometheusConfig.marketCapChange24Relative)
	prometheusConfig.registry.MustRegister(prometheusConfig.high24)
	prometheusConfig.registry.MustRegister(prometheusConfig.low24)
}

func setupWebserver() {
	// Register prom metrics path in http serv
	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		prometheusConfig.registry,
		promhttp.HandlerFor(prometheusConfig.registry, promhttp.HandlerOpts{}),
	))

	// Init & start serv
	prometheusConfig.httpServ = &http.Server{
		Addr:    fmt.Sprintf(":%d", prometheusConfig.httpServerPort),
		Handler: httpMux,
	}
	go func() {
		log.Infof("> Starting HTTP server at %s", prometheusConfig.httpServ.Addr)
		err := prometheusConfig.httpServ.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Errorf("HTTP Server errored out %v", err)
		}
	}()

}
