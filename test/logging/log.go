package logging

import (
	"bufio"
	"io"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

type LogRecorder struct {
	logs []string
}

func (l *LogRecorder) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	l.logs = append(l.logs, msg)
}

func (l *LogRecorder) AssertHasString(require *require.Assertions, s string) {
	for _, log := range l.logs {
		if strings.Contains(log, s) {
			return
		}
	}

	require.Fail("Logs do not contain expected string: "+s, l.logs)
}

// hijackLogs hijacks the log output to capture it
// it removes the hooks after the test is done and returns the logger to previous state
func HijackLogs() (*LogRecorder, func()) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	recorder := &LogRecorder{}
	newLogger := log.Hook(recorder)

	// The new logger should write to /dev/null
	newLogger = newLogger.Output(zerolog.ConsoleWriter{Out: bufio.NewWriter(io.Discard)})

	// Swap the logger
	previousLogger := log.Logger
	log.Logger = newLogger

	return recorder, func() {
		log.Logger = previousLogger
	}
}
