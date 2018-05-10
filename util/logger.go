// logger

package util

import "go.uber.org/zap"

const (
	DebugLevelEtcd uint64 = 1 << iota
	DebugLevelApi
	DebugLevelNotify
)

var slogger *zap.SugaredLogger
var debugLevel uint64

func init() {
	logger, _ := zap.NewDevelopment(zap.AddCallerSkip(1))
	defer logger.Sync()
	slogger = logger.Sugar()
}

func DebugLog(lvl uint64, msg string, keysAndValues ...interface{}) {
	if debugLevel&lvl == 0 {
		return
	}
	slogger.Infow(msg, keysAndValues...)
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

func SetDebugLevel(lvl uint64) {
	debugLevel |= lvl
}

func ClearDebugLevel(lvl uint64) {
	debugLevel &= ^lvl
}

func GetDebugLevel() uint64 {
	return debugLevel
}
