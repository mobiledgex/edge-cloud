// logger

package log

import (
	"context"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var slogger *zap.SugaredLogger
var debugLevel uint64
var mux sync.Mutex

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

func SetDebugLevelStrs(list string) {
	strs := strings.Split(list, ",")
	for _, str := range strs {
		val, ok := DebugLevel_value[str]
		if ok {
			SetDebugLevelEnum(DebugLevel(val))
		}
	}
}

type Api struct{}

func (*Api) EnableDebugLevels(ctx context.Context, lvls *DebugLevels) (*DebugResult, error) {
	SetDebugLevels(lvls.Levels)
	return &DebugResult{}, nil
}

func (*Api) DisableDebugLevels(ctx context.Context, lvls *DebugLevels) (*DebugResult, error) {
	ClearDebugLevels(lvls.Levels)
	return &DebugResult{}, nil
}

func (*Api) ShowDebugLevels(ctx context.Context, in *DebugLevels) (*DebugLevels, error) {
	lvls := DebugLevels{}
	lvls.Levels = make([]DebugLevel, 0)
	for ii := 0; ii < 64; ii++ {
		if (debugLevel & (uint64(1) << uint(ii))) != 0 {
			lvls.Levels = append(lvls.Levels, DebugLevel(ii))
		}
	}
	return &lvls, nil
}
