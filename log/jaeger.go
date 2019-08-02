package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

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
var lastLog map[string]time.Time
var lastLogMux sync.Mutex
var SpanServiceName string

// Use JAEGER_AGEN_HOST and JAEGER_AGENT_PORT to send UDP traces
// to different host:port (otherwise uses localhost:6831).
func InitTracer() {
	lastLog = make(map[string]time.Time)

	SpanServiceName = filepath.Base(os.Args[0])
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 0.001,
		},
		Reporter: &config.ReporterConfig{},
	}
	t, closer, err := cfg.New(SpanServiceName, config.Logger(zap.NewLogger(slogger.Desugar())))
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
