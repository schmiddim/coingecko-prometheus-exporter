package main

import (
	"context"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"io"
)

func initTracing() context.Context {
	// Jaeger Tracing Tutorial https://github.com/yurishkuro/opentracing-tutorial/tree/master/go/lesson03
	cfg := jaegerCfg.Configuration{
		ServiceName: "coingecko-exporter",
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans: true,
		},
	}
	jMetricsFactory := metrics.NullFactory
	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegerCfg.Logger(jaegerLog.NullLogger),
		jaegerCfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Fatal("fuck ", err)
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			log.Fatal("Jaeger", err)
		}
	}(closer)
	span := tracer.StartSpan("main")
	defer span.Finish()
	return opentracing.ContextWithSpan(context.Background(), span)

}
