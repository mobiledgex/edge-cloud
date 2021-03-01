package log

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics/metricstest"
)

var ReporterMetrics *metricstest.Backend

// Adapted from config.NewReporter, in order to be able to pass
// in TLS config to http transport.
// If the jaeger client code
// could expose transport (or tls.Config) in their Configuration,
// then we could avoid this duplication of NewReporter.
func NewReporter(serviceName string, tlsConfig *tls.Config, rc *config.ReporterConfig, logger jaeger.Logger) jaeger.Reporter {
	opts := make([]transport.HTTPOption, 0)
	opts = append(opts, transport.HTTPBatchSize(1))
	if tlsConfig != nil {
		trans := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		opts = append(opts, transport.HTTPRoundTripper(trans))
	}
	factory := metricstest.NewFactory(0)
	metrics := jaeger.NewMetrics(factory, map[string]string{})
	sender := transport.NewHTTPTransport(rc.CollectorEndpoint, opts...)
	reporter := jaeger.NewRemoteReporter(
		sender,
		jaeger.ReporterOptions.QueueSize(rc.QueueSize),
		jaeger.ReporterOptions.BufferFlushInterval(rc.BufferFlushInterval),
		jaeger.ReporterOptions.Metrics(metrics),
		jaeger.ReporterOptions.Logger(logger))
	if rc.LogSpans && logger != nil {
		reporter = jaeger.NewCompositeReporter(jaeger.NewLoggingReporter(logger), reporter)
	}
	ReporterMetrics = factory.Backend
	go reportMetrics(factory.Backend)
	return reporter
}

func reportMetrics(metrics *metricstest.Backend) {
	for {
		time.Sleep(5 * time.Minute)
		counters, gauges := metrics.Snapshot()
		span := StartSpan(DebugLevelInfo, "reporter metrics")
		ctx := ContextWithSpan(context.Background(), span)
		SpanLog(ctx, DebugLevelInfo, "reporter metrics", "counters", counters, "gauges", gauges)
		span.Finish()
	}
}
