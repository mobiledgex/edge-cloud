// logger

package log

import (
	"strings"
	"sync"

	"go.uber.org/zap"
)

var slogger *zap.SugaredLogger
var debugLevel uint64
var mux sync.Mutex
var spanlogger *zap.Logger

func init() {
	logger, _ := zap.NewDevelopment(zap.AddCallerSkip(1))
	defer logger.Sync()
	slogger = logger.Sugar()

	// logger that does not log caller, optimization for
	// span logging to local disk since caller is derived by spanlog.
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	spanlogger, _ = cfg.Build()
	defer spanlogger.Sync()
}

// Deprecated: you should use SpanLog instead.
func DebugLog(lvl uint64, msg string, keysAndValues ...interface{}) {
	if debugLevel&lvl == 0 && lvl != DebugLevelInfo {
		return
	}
	slogger.Infow(msg, keysAndValues...)
}

func DebugSLog(slog *zap.SugaredLogger, lvl uint64, msg string, keysAndValues ...interface{}) {
	if debugLevel&lvl == 0 {
		return
	}
	slog.Infow(msg, keysAndValues...)
}

func InfoLog(msg string, keysAndValues ...interface{}) {
	slogger.Infow(msg, keysAndValues...)
}

func WarnLog(msg string, keysAndValues ...interface{}) {
	slogger.Warnw(msg, keysAndValues...)
}

func FatalLog(msg string, keysAndValues ...interface{}) {
	slogger.Fatalw(msg, keysAndValues...)
}

func AuditLogStart(id uint64, cmd, client, user string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0)
	args = append(args, "id")
	args = append(args, id)
	args = append(args, "cmd")
	args = append(args, cmd)
	args = append(args, "client")
	args = append(args, client)
	args = append(args, "user")
	args = append(args, user)
	args = append(args, keysAndValues...)
	slogger.Infow("Audit start", args...)
}

func AuditLogEnd(id uint64, err error) {
	res := "success"
	msg := ""
	if err != nil {
		res = "failure"
		msg = err.Error()
	}
	slogger.Infow("Audit end", "id", id, "result", res, "msg", msg)
}

func enumToBit(in DebugLevel) uint64 {
	return uint64(1) << uint(in)
}

func SetDebugLevel(lvl uint64) {
	mux.Lock()
	defer mux.Unlock()
	debugLevel |= lvl
}

func ClearDebugLevel(lvl uint64) {
	mux.Lock()
	defer mux.Unlock()
	debugLevel &= ^lvl
}

func GetDebugLevel() uint64 {
	return debugLevel
}

func SetDebugLevelEnum(val DebugLevel) {
	SetDebugLevel(enumToBit(val))
}

func ClearDebugLevelEnum(val DebugLevel) {
	ClearDebugLevel(enumToBit(val))
}

func SetDebugLevels(vals []DebugLevel) {
	for _, val := range vals {
		SetDebugLevelEnum(val)
	}
}

func ClearDebugLevels(vals []DebugLevel) {
	for _, val := range vals {
		ClearDebugLevelEnum(val)
	}
}

func stringToDebugLevel(val string) (DebugLevel, bool) {
	val = strings.ToLower(strings.TrimSpace(val))
	lvl, ok := DebugLevel_value[val]
	return DebugLevel(lvl), ok
}

func SetDebugLevelStrs(list string) {
	strs := strings.Split(list, ",")
	for _, str := range strs {
		val, ok := stringToDebugLevel(str)
		if ok {
			SetDebugLevelEnum(val)
		}
	}
}

func ClearDebugLevelStrs(list string) {
	strs := strings.Split(list, ",")
	for _, str := range strs {
		val, ok := stringToDebugLevel(str)
		if ok {
			ClearDebugLevelEnum(val)
		}
	}
}

func DebugLevels() []DebugLevel {
	lvls := make([]DebugLevel, 0)
	for ii := 0; ii < 64; ii++ {
		if (debugLevel & (uint64(1) << uint(ii))) != 0 {
			lvls = append(lvls, DebugLevel(ii))
		}
	}
	return lvls
}

func GetDebugLevelStrs() string {
	lvls := make([]string, 0)
	for ii := 0; ii < 64; ii++ {
		if (debugLevel & (uint64(1) << uint(ii))) != 0 {
			lvls = append(lvls, DebugLevel_name[int32(ii)])
		}
	}
	return strings.Join(lvls, ",")
}
