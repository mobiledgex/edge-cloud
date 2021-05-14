package log

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"

	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics/metricstest"
)

var ReporterMetrics *metricstest.Backend

type ReporterCloser struct {
	reporter jaeger.Reporter
	done     chan struct{}
}

func (s *ReporterCloser) Close() error {
	s.reporter.Close()
	close(s.done)
	return nil
}

// Adapted from config.NewReporter, in order to be able to pass
// in TLS config to http transport.
// If the jaeger client code
// could expose transport (or tls.Config) in their Configuration,
// then we could avoid this duplication of NewReporter.
func NewReporter(serviceName string, tlsConfig *tls.Config, rc *config.ReporterConfig, logger jaeger.Logger) (jaeger.Reporter, io.Closer) {
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
	closer := &ReporterCloser{}
	closer.reporter = reporter
	closer.done = make(chan struct{})
	go reportMetrics(factory.Backend, closer.done)
	return reporter, closer
}

func reportMetrics(metrics *metricstest.Backend, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-time.After(5 * time.Minute):
		}
		counters, gauges := metrics.Snapshot()
		span := StartSpan(DebugLevelInfo, "reporter metrics")
		ctx := ContextWithSpan(context.Background(), span)
		SpanLog(ctx, DebugLevelInfo, "reporter metrics", "counters", counters, "gauges", gauges)
		span.Finish()
	}
}
