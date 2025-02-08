package go_bat

import (
	"fmt"
	"github.com/phsym/console-slog"
	"log/slog"
	"os"
)

type LoggerOutputType string

const (
	LoggerOutputTypeHuman LoggerOutputType = "human"
	LoggerOutputTypeJSON  LoggerOutputType = "json"
)

func NewLogger(outputType LoggerOutputType, opts *slog.HandlerOptions, noColor bool) *Logger {
	switch outputType {
	case LoggerOutputTypeJSON:
		return &Logger{slog.New(slog.NewJSONHandler(os.Stdout, opts))}
	default:
		hOpts := &console.HandlerOptions{
			AddSource: opts.AddSource,
			Level:     opts.Level,
			NoColor:   noColor,
		}
		return &Logger{slog.New(console.NewHandler(os.Stdout, hOpts))}
	}
}

type Logger struct {
	*slog.Logger
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v))
}

func (l *Logger) Println(v ...interface{}) {
	l.Info(fmt.Sprint(v))
}
