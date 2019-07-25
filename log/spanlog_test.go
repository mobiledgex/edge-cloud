// +build performance

// run this test by using "go test -tags=performance"
package log

import (
	fmt "fmt"
	"os"
	"testing"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	context "golang.org/x/net/context"
)

func TestLoggingSpeed(t *testing.T) {
	/*
		docker run --rm --name jaeger \
		  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
		  -p 5775:5775/udp \
		  -p 6831:6831/udp \
		  -p 6832:6832/udp \
		  -p 5778:5778 \
		  -p 16686:16686 \
		  -p 14268:14268 \
		  -p 9411:9411 \
		  jaegertracing/all-in-one:1.12
	*/
	SetDebugLevelStrs("api")
	file := "/tmp/testout.log"
	os.Remove(file)
	_, err := os.Create(file)
	require.Nil(t, err)

	// send logs to file
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{file}
	logger, err := cfg.Build()
	require.Nil(t, err)
	slogger = logger.Sugar()
	logger.Sync()

	InitTracer()
	defer FinishTracer()

	limit := 999999

	start := time.Now()
	for i := 0; i < limit; i++ {
		DebugLog(DebugLevelApi, "some message1",
			"tag1", "tag1", "tag2", "tag2",
			"val1", "val1", "val2", "val2")
		DebugLog(DebugLevelApi, "some message2",
			"tag1", "tag1", "tag2", "tag2",
			"val1", "val1", "val2", "val2")
		// It's debateable if this should be 2 logs
		// or 4, given that span logging is effectively
		// span start + 2 logs + span end, which ends up
		// being 4 data points, but really only the 2 logs are actually
		// written to local disk.
		/*
			DebugLog(DebugLevelApi, "some message3",
				"tag1", "tag1", "tag2", "tag2",
				"val1", "val1", "val2", "val2")
			DebugLog(DebugLevelApi, "some message4",
				"tag1", "tag1", "tag2", "tag2",
				"val1", "val1", "val2", "val2")
		*/
	}
	elapsed := time.Since(start)
	fmt.Printf("log time took %s\n", elapsed)

	// send logs to file
	os.Remove(file)
	_, err = os.Create(file)
	require.Nil(t, err)
	cfg.DisableCaller = true
	spanlogger, err = cfg.Build()
	require.Nil(t, err)
	spanlogger.Sync()

	start = time.Now()
	for i := 0; i < limit; i++ {
		span := StartSpan(DebugLevelApi, "test")
		span.SetTag("tag1", "tag1")
		span.SetTag("tag2", "tag2")
		ctx := opentracing.ContextWithSpan(context.Background(), span)

		SpanLog(ctx, DebugLevelApi, "some message1", "val1", "val1", "val2", "val2")
		SpanLog(ctx, DebugLevelApi, "some message2", "val1", "val1", "val2", "val2")
		span.Finish()
	}
	elapsed = time.Since(start)
	fmt.Printf("span log time took %s\n", elapsed)

	// Note the second test both writes to disk and writes to
	// jaeger, so it is doing the same work plus more.
	// The things that appear to take the most time are:
	// writing to disk, and computing the current file via runtime.Caller().
	// Test results show an almost 1x overhead for adding on span logging,
	// which is not too bad considering we should not be logging frequently.
	//
	// go test -run TestLoggingSpeed
	// log time took 31.009287796s
	// span log time took 51.11326687s
	// PASS
	// ok  	github.com/mobiledgex/edge-cloud/log	82.191s

}
