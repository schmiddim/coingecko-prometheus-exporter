package main

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type prometheusConfigStruct struct {
	registry       *prometheus.Registry
	vectors        []*prometheus.GaugeVec
	httpServerPort uint
	httpServ       *http.Server
	updateInterval time.Duration
	debug          bool
	configFile     string
	currency       string
	gaugeVectors   map[string]*prometheus.GaugeVec
}

var prometheusConfig = prometheusConfigStruct{
	gaugeVectors: make(map[string]*prometheus.GaugeVec),
	registry:     prometheus.NewRegistry(),
}

func setupGauges(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "setupGauges")
	defer span.Finish()

	gaugeNames := []string{

		"currentPrice",
		"ath",
		"athRelative",
		"change1h",
		"change24h",
		"change7d",
		"change14d",
		"change30d",
		"change60d",
		"change200d",
		"marketCapChange24Relative",
		"marketCap",
		"high24",
		"low24",
	}
	for _, name := range gaugeNames {

		prometheusConfig.gaugeVectors[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "coin_gecko",
			Name:      name,
			Help:      name}, []string{"symbol"})
		prometheusConfig.registry.MustRegister(prometheusConfig.gaugeVectors[name])
	}

}

func setupWebserver(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "setupWebserver")
	defer span.Finish()

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
