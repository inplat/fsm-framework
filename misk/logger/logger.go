package logger

import (
	"context"
	"io"
	"os"

	zlog "github.com/rs/zerolog"
)

const MessageFieldName = "msg"

var globalLogWriter io.Writer

type Logger struct {
	zlog.Logger
}

type loggerCtxKey struct{}

func (l Logger) Copy() *Logger {
	return &Logger{
		Logger: l.Logger.With().Logger(), // With() создаёт копию логгера
	}
}

func FromLogger(zl zlog.Logger) *Logger {
	return Logger{
		Logger: zl,
	}.Copy()
}

// New создаёт новый объект Logger с записью в io.Writer, установленный методами Init*Logger
func New() *Logger {
	return &Logger{
		Logger: initLogger().Logger(),
	}
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	if lp, ok := ctx.Value(loggerCtxKey{}).(*Logger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	} else if l.GetLevel() == zlog.Disabled {
		// Do not store disabled logger.
		return ctx
	}

	return context.WithValue(l.Logger.WithContext(ctx), loggerCtxKey{}, l)
}

func Ctx(ctx context.Context) *Logger {
	var logger *Logger
	if l, ok := ctx.Value(loggerCtxKey{}).(*Logger); ok {
		logger = l
	} else {
		logger = FromLogger(*zlog.Ctx(ctx))
	}

	return logger
}

func initLogger() zlog.Context {
	zlog.MessageFieldName = MessageFieldName

	return zlog.New(globalLogWriter).With().Int("pid", os.Getpid()).Timestamp()
}
