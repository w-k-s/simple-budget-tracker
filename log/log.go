package log

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	WithContext(ctx context.Context) context.Context
	WithFields(map[string]interface{}) Logger
	Info() Event
	Err(err error) Event
	Print(msg string)
	Printf(msg string, v ...interface{})
	Fatal(msg string)
}

type Event interface {
	Str(key string, value string) Event
	Int32(key string, value int32) Event
	Int64(key string, value int64) Event
	UInt64(key string, value uint64) Event
	Duration(key string, value time.Duration) Event
	Struct(key string, value interface{}) Event
	Msg(msg string)
	Msgf(format string, args ...interface{})
}

func NewLogger(w io.Writer) Logger {
	l := zerolog.New(w)
	return &internalLogger{
		l: &l,
	}
}

func Info() Event {
	return &internalLogEvent{log.Info()}
}

func InfoCtx(ctx context.Context) Event {
	if logger := withLogger(ctx); logger != nil {
		return logger.Info()
	} else {
		return Info()
	}
}

func Err(err error) Event {
	return &internalLogEvent{log.Err(err)}
}

func ErrCtx(ctx context.Context, err error) Event {
	if logger := withLogger(ctx); logger != nil {
		return logger.Err(err)
	} else {
		return Err(err)
	}
}

func Print(msg string) {
	log.Info().Msg(msg)
}

func Printf(msg string, v ...interface{}) {
	log.Info().Msgf(msg, v...)
}

func Fatal(err error) {
	log.Fatal().Err(err)
}

func Fatalf(msg string, v ...interface{}) {
	log.Fatal().Msgf(msg, v...)
}

func WithFields(fields map[string]interface{}) Logger {
	l := log.With().Fields(fields).Logger()
	return &internalLogger{&l}
}

func withLogger(ctx context.Context) Logger {
	return &internalLogger{log.Ctx(ctx)}
}

// --
type internalLogger struct {
	l *zerolog.Logger
}

func (l *internalLogger) WithContext(ctx context.Context) context.Context {
	return l.l.WithContext(ctx)
}

func (l *internalLogger) WithFields(fields map[string]interface{}) Logger {
	child := l.l.With().Fields(fields).Logger()
	return &internalLogger{&child}
}

func (l *internalLogger) Info() Event {
	return &internalLogEvent{l.l.Info()}
}

func (l *internalLogger) Err(err error) Event {
	return &internalLogEvent{l.l.Err(err)}
}

func (l *internalLogger) Print(msg string) {
	l.l.Info().Msg(msg)
}

func (l *internalLogger) Printf(msg string, v ...interface{}) {
	l.l.Info().Msgf(msg, v...)
}

func (l *internalLogger) Fatal(msg string) {
	l.l.Fatal().Msg(msg)
}

type internalLogEvent struct {
	e *zerolog.Event
}

func (e *internalLogEvent) Str(key string, value string) Event {
	return &internalLogEvent{e.e.Str(key, value)}
}

func (e *internalLogEvent) Int32(key string, value int32) Event {
	return &internalLogEvent{e.e.Int32(key, value)}
}

func (e *internalLogEvent) Int64(key string, value int64) Event {
	return &internalLogEvent{e.e.Int64(key, value)}
}

func (e *internalLogEvent) UInt64(key string, value uint64) Event {
	return &internalLogEvent{e.e.Uint64(key, value)}
}

func (e *internalLogEvent) Duration(key string, value time.Duration) Event {
	return &internalLogEvent{e.e.Dur(key, value)}
}

func (e *internalLogEvent) Struct(key string, value interface{}) Event {
	return &internalLogEvent{e.e.Interface(key, value)}
}

func (e *internalLogEvent) Msg(msg string) {
	e.e.Msg(msg)
}

func (e *internalLogEvent) Msgf(format string, args ...interface{}) {
	e.e.Msgf(format, args...)
}
