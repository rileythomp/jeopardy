package log

import (
	"fmt"
	"log"
	"os"
)

func Debug(s string, args ...any) {
	log.Printf("[DEBUG] %s\n", fmt.Sprintf(s, args...))
}

func Infof(s string, args ...any) {
	log.Printf("[INF0] %s\n", fmt.Sprintf(s, args...))
}

func Warnf(s string, args ...any) {
	log.Printf("[WARN] %s\n", fmt.Sprintf(s, args...))
}

func Errorf(s string, args ...any) {
	log.Printf("[ERROR] %s\n", fmt.Sprintf(s, args...))
}

func Fatalf(s string, args ...any) {
	log.Printf("[FATAL] %s\n", fmt.Sprintf(s, args...))
	os.Exit(1)
}

func Panicf(s string, args ...any) {
	log.Printf("[PANIC] %s\n", fmt.Sprintf(s, args...))
	panic(fmt.Sprintf(s, args...))
}

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"strings"

// 	kitLog "github.com/go-kit/log"
// 	kitLevel "github.com/go-kit/log/level"
// )

// const (
// 	instanceCallerDepth  = 7
// 	singletonCallerDepth = 8
// 	LogSizeLimit         = 100000
// )

// type Format int

// const (
// 	JSON Format = iota
// 	Logfmt
// )

// type Level int

// const (
// 	NoneLevel Level = iota
// 	ErrorLevel
// 	WarnLevel
// 	InfoLevel
// 	DebugLevel
// )

// type Options struct {
// 	callerDepth int
// 	Name        string
// 	Environment string
// 	Region      string
// 	Level       string
// 	Format      Format
// 	Writer      io.Writer
// }

// func stringToLevel(level string) Level {
// 	switch strings.ToLower(level) {
// 	case "none":
// 		return NoneLevel
// 	case "error":
// 		return ErrorLevel
// 	case "warn":
// 		return WarnLevel
// 	case "info":
// 		return InfoLevel
// 	case "debug":
// 		return DebugLevel
// 	default:
// 		return InfoLevel
// 	}
// }

// func createBaseLogger(opts Options) kitLog.Logger {
// 	var base kitLog.Logger

// 	if opts.Writer == nil {
// 		opts.Writer = os.Stdout
// 	}

// 	switch opts.Format {
// 	case Logfmt:
// 		base = kitLog.NewLogfmtLogger(opts.Writer)
// 	case JSON:
// 		fallthrough
// 	default:
// 		base = kitLog.NewJSONLogger(opts.Writer)
// 	}

// 	// This is not required since SwapLogger uses a SyncLogger and can be used concurrently
// 	// base = kitLog.NewSyncLogger(base)

// 	if opts.callerDepth == 0 {
// 		opts.callerDepth = instanceCallerDepth
// 	}

// 	base = kitLog.With(base,
// 		"caller", kitLog.Caller(opts.callerDepth),
// 		"timestamp", kitLog.DefaultTimestampUTC,
// 	)

// 	if opts.Name != "" {
// 		base = kitLog.With(base, "logger", opts.Name)
// 	}

// 	if opts.Environment != "" {
// 		base = kitLog.With(base, "environment", opts.Environment)
// 	}

// 	if opts.Region != "" {
// 		base = kitLog.With(base, "region", opts.Region)
// 	}

// 	return base
// }

// func createFilteredLogger(base kitLog.Logger, level Level) kitLog.Logger {
// 	var filtered kitLog.Logger

// 	switch level {
// 	case NoneLevel:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowNone())
// 	case ErrorLevel:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowError())
// 	case WarnLevel:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowWarn())
// 	case InfoLevel:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowInfo())
// 	case DebugLevel:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowDebug())
// 	default:
// 		filtered = kitLevel.NewFilter(base, kitLevel.AllowInfo())
// 	}

// 	return filtered
// }

// type Logger struct {
// 	Level  Level
// 	base   kitLog.Logger
// 	logger *kitLog.SwapLogger
// }

// func NewLogger(opts Options) *Logger {
// 	level := stringToLevel(opts.Level)
// 	base := createBaseLogger(opts)
// 	filtered := createFilteredLogger(base, level)

// 	logger := new(kitLog.SwapLogger)
// 	logger.Swap(filtered)

// 	return &Logger{
// 		Level:  level,
// 		base:   base,
// 		logger: logger,
// 	}
// }

// func NewVoidLogger() *Logger {
// 	nop := kitLog.NewNopLogger()

// 	logger := new(kitLog.SwapLogger)
// 	logger.Swap(nop)

// 	return &Logger{
// 		base:   nop,
// 		logger: logger,
// 	}
// }

// func (l *Logger) With(kv ...interface{}) *Logger {
// 	level := l.Level
// 	base := kitLog.With(l.base, kv...)
// 	filtered := createFilteredLogger(base, level)

// 	logger := new(kitLog.SwapLogger)
// 	logger.Swap(filtered)

// 	return &Logger{
// 		Level:  level,
// 		base:   base,
// 		logger: logger,
// 	}
// }

// func (l *Logger) SetLevel(level string) {
// 	l.Level = stringToLevel(level)
// 	l.logger.Swap(createFilteredLogger(l.base, l.Level))
// }

// func (l *Logger) SetOptions(opts Options) {
// 	l.Level = stringToLevel(opts.Level)
// 	l.base = createBaseLogger(opts)
// 	l.logger.Swap(createFilteredLogger(l.base, l.Level))
// }

// const choppedOff = "...ChoppedOff..."

// func (l *Logger) Debug(message string) {
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}

// 	_ = kitLevel.Debug(l.logger).Log("message", message)

// }

// func (l *Logger) Debugf(format string, v ...interface{}) {
// 	message := fmt.Sprintf(format, v...)
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}

// 	_ = kitLevel.Debug(l.logger).Log("message", message)

// }

// func (l *Logger) Info(message string) {
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Info(l.logger).Log("message", message)
// }

// func (l *Logger) Infof(format string, v ...interface{}) {
// 	message := fmt.Sprintf(format, v...)
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Info(l.logger).Log("message", message)
// }

// func (l *Logger) Warn(message string) {
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Warn(l.logger).Log("message", message)
// }

// func (l *Logger) Warnf(format string, v ...interface{}) {
// 	message := fmt.Sprintf(format, v...)
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Warn(l.logger).Log("message", message)
// }

// func (l *Logger) Error(message string) {
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Error(l.logger).Log("message", message)
// }

// func (l *Logger) Errorf(format string, v ...interface{}) {
// 	message := fmt.Sprintf(format, v...)
// 	if len(message) > LogSizeLimit {
// 		message = message[0:LogSizeLimit-20] + choppedOff
// 	}
// 	_ = kitLevel.Error(l.logger).Log("message", message)
// }

// var singleton = NewLogger(Options{
// 	Name:        "singleton",
// 	callerDepth: singletonCallerDepth,
// })

// func SetLevel(level string) {
// 	singleton.SetLevel(level)
// }

// func SetOptions(opts Options) {
// 	opts.callerDepth = 8
// 	singleton.SetOptions(opts)
// }

// func Debug(message string) {
// 	singleton.Debug(message)
// }

// func Debugf(format string, v ...interface{}) {
// 	singleton.Debugf(format, v...)
// }

// func Info(message string) {
// 	singleton.Info(message)
// }

// func Infof(format string, v ...interface{}) {
// 	singleton.Infof(format, v...)
// }

// func Warn(message string) {
// 	singleton.Warn(message)
// }

// func Warnf(format string, v ...interface{}) {
// 	singleton.Warnf(format, v...)
// }

// func Error(message string) {
// 	singleton.Error(message)
// }

// func Errorf(format string, v ...interface{}) {
// 	singleton.Errorf(format, v...)
// }

// type contextKey string

// const loggerContextKey = contextKey("logger")

// func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
// 	return context.WithValue(ctx, loggerContextKey, logger)
// }

// func LoggerFromContext(ctx context.Context) *Logger {
// 	val := ctx.Value(loggerContextKey)
// 	if logger, ok := val.(*Logger); ok {
// 		return logger
// 	}

// 	return singleton
// }
