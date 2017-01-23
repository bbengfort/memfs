// Mechanism for handling application level logging

package memfs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//===========================================================================
// Log Level Type
//===========================================================================

// LogLevel characterizes the severity of the log message.
type LogLevel int

// Severity levels of log messages.
const (
	LevelDebug LogLevel = 1 + iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String representations of the various log levels.
var levelNames = []string{
	"DEBUG", "INFO", "WARN", "ERROR", "FATAL",
}

// String representation of the log level.
func (level LogLevel) String() string {
	return levelNames[level-1]
}

// LevelFromString parses a string and returns the LogLevel
func LevelFromString(level string) LogLevel {
	// Perform string cleanup for matching
	level = strings.ToUpper(level)
	level = strings.Trim(level, " ")

	switch level {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

//===========================================================================
// Logger wrapper for log.Logger and logging initialization methods
//===========================================================================

// Logger wraps the log.Logger to write to a file on demand and to specify a
// miminum severity that is allowed for writing.
type Logger struct {
	Level  LogLevel       // The minimum severity to log to
	logger *log.Logger    // The wrapped logger for concurrent logging
	output io.WriteCloser // Handle to the open log file or writer object
}

// InitLogger creates a Logger object by passing a configuration that contains
// the minimum log level and an optional path to write the log out to.
func InitLogger(path string, level string) (*Logger, error) {
	newLogger := new(Logger)
	newLogger.Level = LevelFromString(level)

	// If a path is specified create a handle to the writer.
	if path != "" {

		var err error
		newLogger.output, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}

	} else {
		newLogger.output = os.Stdout
	}

	newLogger.logger = log.New(newLogger.output, "", 0)

	return newLogger, nil
}

// Close the logger and any open file handles.
func (logger *Logger) Close() error {
	if err := logger.output.Close(); err != nil {
		return err
	}
	return nil
}

// GetHandler returns the io.Writer object that is on the logger.
func (logger *Logger) GetHandler() io.Writer {
	return logger.output
}

// SetHandler sets a new io.WriteCloser object onto the logger
func (logger *Logger) SetHandler(writer io.WriteCloser) {
	logger.output = writer
	logger.logger.SetOutput(writer)
}

//===========================================================================
// Logging handlers
//===========================================================================

// Log a message at the appropriate severity. The Log method behaves as a
// format function, and a layout string can be passed with arguments.
// The current logging format is "%(level)s [%(jsontime)s]: %(message)s"
func (logger *Logger) Log(layout string, level LogLevel, args ...interface{}) {

	// Only log if the log level matches the log request
	if level >= logger.Level {
		msg := fmt.Sprintf(layout, args...)
		msg = fmt.Sprintf("%-7s [%s]: %s", level, time.Now().Format(JSONDateTime), msg)

		// If level is fatal then log fatal.
		if level == LevelFatal {
			logger.logger.Fatalln(msg)
		} else {
			logger.logger.Println(msg)
		}

	}

}

// Debug message helper function
func (logger *Logger) Debug(msg string, args ...interface{}) {
	logger.Log(msg, LevelDebug, args...)
}

// Info message helper function
func (logger *Logger) Info(msg string, args ...interface{}) {
	logger.Log(msg, LevelInfo, args...)
}

// Warn message helper function
func (logger *Logger) Warn(msg string, args ...interface{}) {
	logger.Log(msg, LevelWarn, args...)
}

// Error message helper function
func (logger *Logger) Error(msg string, args ...interface{}) {
	logger.Log(msg, LevelError, args...)
}

// Fatal message helper function
func (logger *Logger) Fatal(msg string, args ...interface{}) {
	logger.Log(msg, LevelFatal, args...)
}

//===========================================================================
// HTTP logging handler for the C2S API and web interface
//===========================================================================

// :method :url :status :response-time ms - :res[content-length]
const webLogFmt = "c2s %s %s %d %s - %d"

// WebLogger is a decorator for http handlers to record HTTP requests using
// the logger API and syntax, which must be passed in as the first argument.
func WebLogger(log *Logger, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &responseLogger{w: w}
		inner.ServeHTTP(lw, r)

		log.Info(webLogFmt, r.Method, r.RequestURI, lw.Status(), time.Since(start), lw.Size())

	})
}

// responseLogger is a wrapper of http.ResponseWriter that keeps track of its
// HTTP status code and body size for reporting to the console.
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}

	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}
