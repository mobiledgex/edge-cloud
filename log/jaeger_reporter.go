package log

import (
	"crypto/tls"
	"net/http"

	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
)

// Adapted from config.NewReporter, in order to be able to pass
// in TLS config to http transport.
// Unfortunately we can't pass metrics in, because jaeger.Metrics
// is created in the NewTracer call. If the jaeger client code
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
	sender := transport.NewHTTPTransport(rc.CollectorEndpoint, opts...)
	reporter := jaeger.NewRemoteReporter(
		sender,
		jaeger.ReporterOptions.QueueSize(rc.QueueSize),
		jaeger.ReporterOptions.BufferFlushInterval(rc.BufferFlushInterval),
		jaeger.ReporterOptions.Logger(logger))
	if rc.LogSpans && logger != nil {
		reporter = jaeger.NewCompositeReporter(jaeger.NewLoggingReporter(logger), reporter)
	}
	return reporter
}
