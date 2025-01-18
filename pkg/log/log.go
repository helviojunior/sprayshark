package log

import (
    "os"
    "fmt"

    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/log"
)

// LLogger is a charmbracelet logger type redefinition
type LLogger = log.Logger

// Logger is this package level logger
var Logger *LLogger

func init() {
    styles := log.DefaultStyles()
    styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
    styles.Values["err"] = lipgloss.NewStyle().Bold(true)

    Logger = log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: false,
    })
    Logger.SetStyles(styles)
    Logger.SetLevel(log.InfoLevel)
}

// EnableDebug enabled debug logging and caller reporting
func EnableDebug() {
    Logger.SetLevel(log.DebugLevel)
    Logger.SetReportCaller(true)
}

// EnableSilence will silence most logs, except this written with Print
func EnableSilence() {
    Logger.SetLevel(log.FatalLevel + 100)
}

// Debug logs debug messages
func Debug(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Debug(msg, keyvals...)
}
func Debugf(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Debug(fmt.Sprintf(format, a...) )
}

// Info logs info messages
func Info(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Info(msg, keyvals...)
}
func Infof(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Info(fmt.Sprintf(format, a...) )
}


// Warn logs warning messages
func Warn(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Warn(msg, keyvals...)
}
func Warnf(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Warn(fmt.Sprintf(format, a...) )
}


// Error logs error messages
func Error(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Error(msg, keyvals...)
}
func Errorf(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Error(fmt.Sprintf(format, a...) )
}

// Fatal logs fatal messages and panics
func Fatal(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Fatal(msg, keyvals...)
}
func Fatalf(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Fatal(fmt.Sprintf(format, a...) )
}


// Print logs messages regardless of level
func Print(msg string, keyvals ...interface{}) {
    Logger.Helper()
    Logger.Print(msg, keyvals...)
}
func Printf(format string, a ...interface{}) {
    Logger.Helper()
    Logger.Print(fmt.Sprintf(format, a...) )
}

// With returns a sublogger with a prefix
func With(keyvals ...interface{}) *LLogger {
    return Logger.With(keyvals...)
}