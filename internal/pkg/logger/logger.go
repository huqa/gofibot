// Package logger defines a logger wrapper for gofibot
package logger

import (
	"context"

	"go.uber.org/zap"
)

type Logger interface {
	WithContext(context.Context) Logger
	With(args ...interface{}) Logger
	Named(string) Logger

	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
}

type LogWrapper struct {
	*zap.SugaredLogger
}

func New(c Configuration) *LogWrapper {
	config := &c.Config

	log, err := config.Build()
	if err != nil {
		Fatal(err)
	}
	zap.RedirectStdLog(log)
	return &LogWrapper{log.Sugar()}
}

// WithContext returns a new logger with some values taken from the context.
func (l *LogWrapper) WithContext(ctx context.Context) Logger {
	newLogger := l.SugaredLogger

	if cid, ok := ctx.Value("correlationID").(string); ok {
		newLogger = newLogger.With(zap.String("correlationID", cid))
	}

	return &LogWrapper{newLogger}
}

// Named returns a new logger with the given name.
// See zap.SugaredLogger.Named for details.
func (l *LogWrapper) Named(name string) Logger {
	return &LogWrapper{l.SugaredLogger.Named(name)}
}

func (l *LogWrapper) With(args ...interface{}) Logger {
	return &LogWrapper{l.SugaredLogger.With(args...)}
}

func Fatal(args ...interface{}) {
	plog, _ := zap.NewProduction()
	log := plog.Sugar()
	log.Fatal(args...)
}
