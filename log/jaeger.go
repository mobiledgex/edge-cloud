package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud/tls"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log/zap"
)

var tracer opentracing.Tracer
var tracerCloser io.Closer

type contextKey struct{}

var spanString = contextKey{}
var noPanicOrphanedSpans bool
var SpanServiceName string

// InitTracer configures the Jaeger OpenTracing client to log traces.
// Set JAEGER_ENDPOINT to http://<jaegerhost>:14268/api/traces to
// specify the Jaeger server.
func InitTracer(tlsCertFile string) {
	SpanServiceName = filepath.Base(os.Args[0])

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}
	ur, err := url.Parse(jaegerEndpoint)
	if err != nil {
		panic(fmt.Sprintf("ERROR: failed to parse jaeger endpoint %s, %v", jaegerEndpoint, err))
	}

	// Set up client-side TLS
	skipVerify := false
	tlsConfig, err := tls.GetTLSClientConfig(ur.Host, tlsCertFile, "", skipVerify)
	if err != nil {
		panic(fmt.Sprintf("ERROR: failed to init TLS client config for cert %s, %v", tlsCertFile, err))
	}
	if tlsConfig == nil {
		ur.Scheme = "http"
	} else {
		ur.Scheme = "https"
	}
	jaegerEndpoint = ur.String()

	// Configure Jaeger client
	// Note that we create the Reporter manually to be able to do mTLS
	rc := &config.ReporterConfig{
		CollectorEndpoint: jaegerEndpoint,
		QueueSize:         500,
	}
	logger := zap.NewLogger(slogger.Desugar())
	reporter := NewReporter(SpanServiceName, tlsConfig, rc, logger)

	cfg := &config.Configuration{
		ServiceName: SpanServiceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 0.001,
		},
	}
	if strings.HasSuffix(os.Args[0], ".test") {
		// unit test, don't bother reporting
		reporter = jaeger.NewNullReporter()
	}

	// Create tracer
	t, closer, err := cfg.NewTracer(
		config.Logger(logger),
		config.Reporter(reporter),
		config.MaxTagValueLength(4096),
	)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	tracer = t
	tracerCloser = closer
	opentracing.SetGlobalTracer(t)

	if _, found := os.LookupEnv("NO_PANIC_ORPHANED_SPANS"); found {
		// suppress the default behavior of panicking on
		// orphaned spans. This can be used for the main
		// deployment to prevent panicking.
		noPanicOrphanedSpans = true
	}
}

func FinishTracer() {
	tracerCloser.Close()
}

// TraceData is used to transport trace/span across boundaries,
// such as via etcd (stored on disk) or notify (cached in keys map).
type TraceData map[string]string

func (t TraceData) Set(key, val string) {
	t[key] = val
}

func (t TraceData) ForeachKey(handler func(key, val string) error) error {
	for k, v := range t {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

func SpanToString(ctx context.Context) string {
	span := SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	var t TraceData
	t = make(map[string]string)
	tracer.Inject(span.Context(), opentracing.TextMap, t)
	val, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(val)
}

func NewSpanFromString(lvl uint64, val, spanName string) opentracing.Span {
	if val != "" {
		var t TraceData
		t = make(map[string]string)
		err := json.Unmarshal([]byte(val), &t)
		if err == nil {
			spanCtx, err := tracer.Extract(opentracing.TextMap, t)
			if err == nil {
				// parent span exists so new lvl is ignored,
				// lvl used for parent is honored.
				return StartSpan(IgnoreLvl, spanName, ext.RPCServerOption(spanCtx))
			}
		}
	}
	return StartSpan(lvl, spanName)
}
