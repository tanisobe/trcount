package trmon

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

func NewLogger(isDebug bool, w io.Writer) *Logger {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	output := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s\t", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("\t%s\t", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\t%s\t", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("\t%s\t", i))
	}
	l := zerolog.New(output).With().Timestamp().Caller().Logger()
	return &Logger{Logger: &l}
}
