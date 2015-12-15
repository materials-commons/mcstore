package app

import (
	"fmt"
	"os"

	"github.com/inconshreveable/log15"
)

// Logger holds the application logger
type Logger struct {
	log15.Logger
}

var (
	// Log Is the global log variable.
	Log = newLog()

	// stdHandler is the log handler with level applied
	stdHandler = log15.StreamHandler(os.Stdout, log15.LogfmtFormat())

	// Default handler used in the package.
	defaultHandler log15.Handler
)

func newLog() *Logger {
	return &Logger{
		Logger: log15.New(),
	}
}

func init() {
	SetDefaultLogHandler(log15.LvlFilterHandler(log15.LvlInfo, stdHandler))
	Log.SetHandler(defaultHandler)
}

// NewLog creates a new instance of the logger using the current default handler
// for its output.
func NewLog(ctx ...interface{}) *Logger {
	l := log15.New(ctx...)
	l.SetHandler(defaultHandler)
	return &Logger{Logger: l}
}

// Errorf will write a formatted Error to the default log.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Debugf will write a formatted Debug to the default log.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Critf will write a formatted Crit to the default log.
func (l *Logger) Critf(format string, args ...interface{}) {
	l.Crit(fmt.Sprintf(format, args...))
}

// Infof will write a formatted Info to the default log.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf will write a formatted Warn to the default log.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Crit(fmt.Sprintf(format, args...))
	Panicf(format, args...)
}

func (l *Logger) Exitf(format string, args ...interface{}) {
	l.Crit(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Logf is short hand to create a message string using fmt.Sprintf.
func Logf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// SetDefaultLogHandler sets the handler for the logger. It wraps handlers in a SyncHandler. You
// should not pass in handlers that are already wrapped in a SyncHandler.
func SetDefaultLogHandler(handler log15.Handler) {
	defaultHandler = log15.SyncHandler(handler)
	Log.SetHandler(defaultHandler)
}

// Sets a new log level for the global logging and the default handler.
func SetLogLvl(lvl log15.Lvl) {
	SetDefaultLogHandler(log15.LvlFilterHandler(lvl, stdHandler))
	Log.SetHandler(defaultHandler)
}

// DefaultLogHandler returns the current handler. It can be used to create additional
// logger instances that all use the same handler for output.
func DefaultLogHandler() log15.Handler {
	return defaultHandler
}
