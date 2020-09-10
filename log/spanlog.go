package log

import (
	"context"
	"runtime"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	trlog "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wrap Span so we can override Finish()
type Span struct {
	*jaeger.Span
	suppress bool // ignore log for show commands etc
}

var IgnoreLvl uint64 = 99999
var SuppressLvl uint64 = 99998
var SamplingEnabled = true

func StartSpan(lvl uint64, operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	if tracer == nil {
		panic("tracer not initialized. Use log.InitTracer()")
	}
	ospan := tracer.StartSpan(operationName, opts...)
	if lvl == SuppressLvl {
		// log to span but not to disk, allows caller to decide
		// right before Finish whether or not to log the whole thing.
		ext.SamplingPriority.Set(ospan, 1)
	} else if lvl != IgnoreLvl {
		if DebugLevelSampled&lvl != 0 {
			if SamplingEnabled {
				// sampled
			} else {
				// always log
				ext.SamplingPriority.Set(ospan, 1)
			}
		} else if DebugLevelInfo&lvl != 0 || debugLevel&lvl != 0 {
			// always log (note DebugLevelInfo is always logged)
			ext.SamplingPriority.Set(ospan, 1)
		} else {
			// don't log
			ext.SamplingPriority.Set(ospan, 0)
		}
	}

	jspan, ok := ospan.(*jaeger.Span)
	if !ok {
		panic("non-jaeger span not supported")
	}
	span := &Span{Span: jspan}
	if lvl == SuppressLvl {
		span.suppress = true
	}

	if jspan.SpanContext().IsSampled() && !span.suppress {
		ec := zapcore.NewEntryCaller(runtime.Caller(1))
		spanlogger.Info(getSpanMsg(span, ec.TrimmedPath(), "start "+operationName))
	}
	return span
}

func ChildSpan(ctx context.Context, lvl uint64, operationName string) (opentracing.Span, context.Context) {
	span := StartSpan(lvl, operationName, opentracing.ChildOf(SpanFromContext(ctx).Context()))
	return span, ContextWithSpan(context.Background(), span)
}

func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

func SetTags(span opentracing.Span, tags map[string]string) {
	for k, v := range tags {
		span.SetTag(k, v)
	}
}

func GetTags(span opentracing.Span) map[string]interface{} {
	sp, ok := span.(*Span)
	if !ok {
		return make(map[string]interface{})
	}
	return sp.Span.Tags()
}

func SetContextTags(ctx context.Context, tags map[string]string) {
	SetTags(SpanFromContext(ctx), tags)
}

func SpanLog(ctx context.Context, lvl uint64, msg string, keysAndValues ...interface{}) {
	if debugLevel&lvl == 0 && lvl != DebugLevelInfo {
		return
	}
	ospan := opentracing.SpanFromContext(ctx)
	if ospan == nil {
		if noPanicOrphanedSpans {
			ospan = StartSpan(DebugLevelInfo, "orphaned")
			defer ospan.Finish()
		} else {
			panic("no span in context")
		}
	}
	span, ok := ospan.(*Span)
	if !ok {
		panic("non-edge-cloud Span not supported")
	}
	if !span.SpanContext().IsSampled() {
		return
	}

	ec := zapcore.NewEntryCaller(runtime.Caller(1))
	lineno := ec.TrimmedPath()
	fields := []trlog.Field{
		trlog.String("msg", msg),
		trlog.String("lineno", lineno),
	}
	kvfields, err := trlog.InterleavedKVToFields(keysAndValues...)
	if err != nil {
		FatalLog("SpanLog invalid args", "err", err)
	}
	fields = append(fields, kvfields...)
	span.LogFields(fields...)

	// Log to disk as well. Pull tags from span.
	// Unfortunately zap logger and opentracing logger, although
	// both implemented by uber, don't use the same Field struct.
	zfields := getFields(keysAndValues)
	// don't write to log file if deferring log decision
	if !span.suppress {
		spanlogger.Info(getSpanMsg(span, lineno, msg), zfields...)
	}
}

func getFields(args []interface{}) []zap.Field {
	fields := []zap.Field{}
	for i := 0; i < len(args); {
		if i == len(args)-1 {
			panic("odd number of args")
		}
		k, v := args[i], args[i+1]
		// InterleavedKVToFields call ensures even number of args
		// and that key is a string
		if keystr, ok := k.(string); ok {
			fields = append(fields, zap.Any(keystr, v))
		}
		i += 2
	}
	return fields
}

// Convenience function for test routines
func StartTestSpan(ctx context.Context) context.Context {
	span := StartSpan(DebugLevelInfo, "test")
	// ignore span.Finish()
	return opentracing.ContextWithSpan(ctx, span)
}

func (s *Span) Finish() {
	if s.suppress {
		return
	}
	s.Span.Finish()

	jspan := s.Span
	if !jspan.SpanContext().IsSampled() {
		return
	}

	ec := zapcore.NewEntryCaller(runtime.Caller(1))
	lineno := ec.TrimmedPath()

	fields := []zap.Field{}
	for k, v := range jspan.Tags() {
		if k == "span.kind" ||
			k == "sampler.type" ||
			k == "sampler.param" ||
			k == "sampling.priority" {
			continue
		}
		fields = append(fields, zap.Any(k, v))
	}
	msg := getSpanMsg(s, lineno, "finish "+s.OperationName())
	spanlogger.Info(msg, fields...)
}

func Unsuppress(ospan opentracing.Span) {
	s, ok := ospan.(*Span)
	if !ok {
		panic("non-edge-cloud Span not supported")
	}
	s.suppress = false
}

func getSpanMsg(s *Span, lineno, msg string) string {
	traceid := "notrace"
	if s != nil {
		traceid = s.Span.SpanContext().TraceID().String()
	}
	return traceid + "\t" + lineno + "\t" + msg
}

func NoLogSpan(span opentracing.Span) {
	ext.SamplingPriority.Set(span, 0)
}

func ForceLogSpan(span opentracing.Span) {
	ext.SamplingPriority.Set(span, 1)
}
