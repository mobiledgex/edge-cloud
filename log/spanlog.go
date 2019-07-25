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

func StartSpan(lvl uint64, operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	if tracer == nil {
		panic("tracer not initialized. Use log.InitTracer()")
	}
	span := tracer.StartSpan(operationName, opts...)
	if DebugLevelSampled&lvl != 0 {
		// sampled
	} else if DebugLevelInfo&lvl != 0 || debugLevel&lvl != 0 {
		// always log (note DebugLevelInfo is always logged)
		ext.SamplingPriority.Set(span, 1)
	} else {
		// don't log
		ext.SamplingPriority.Set(span, 0)
	}
	return span
}

func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

func SpanLog(ctx context.Context, lvl uint64, msg string, keysAndValues ...interface{}) {
	if debugLevel&lvl == 0 {
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
	span, ok := ospan.(*jaeger.Span)
	if !ok {
		panic("non-jaeger span not supported")
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
	zfields := getFields(span, lineno, keysAndValues)
	spanlogger.Info(msg, zfields...)
}

func getFields(span *jaeger.Span, lineno string, args []interface{}) []zap.Field {
	fields := []zap.Field{
		zap.String("lineno", lineno),
	}
	for k, v := range span.Tags() {
		if k == "span.kind" ||
			k == "sampler.type" ||
			k == "sampler.param" ||
			k == "sampling.priority" {
			continue
		}
		fields = append(fields, zap.Any(k, v))
	}
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
