package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// Logs go to stdout only; collecting and rotating them is the job of the
// container runtime or process supervisor.
func init() {
	logger.Level = logrus.InfoLevel
	logger.Formatter = &formatter{}
	logger.SetOutput(os.Stdout)
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level logrus.Level) {
	logger.Level = level
}

// Configure applies LOG_LEVEL (debug, info, warn, error) and LOG_FORMAT
// ("json" or "text"). Empty values keep the current setting.
func Configure(level, format string) error {
	if level != "" {
		lvl, err := logrus.ParseLevel(level)
		if err != nil {
			return fmt.Errorf("invalid LOG_LEVEL %q: %w", level, err)
		}
		logger.Level = lvl
	}
	switch strings.ToLower(format) {
	case "":
	case "json":
		logger.Formatter = &logrus.JSONFormatter{TimestampFormat: time.RFC3339}
	case "text":
		logger.Formatter = &formatter{}
	default:
		return fmt.Errorf("invalid LOG_FORMAT %q: want json or text", format)
	}
	return nil
}

// SetOutput redirects log output, mainly for tests.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// Fields type alias for logrus.Fields
type Fields logrus.Fields

// Debugf logs a message at level Debug.
func Debugf(format string, args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		// Create a log entry with empty fields
		entry := logger.WithFields(logrus.Fields{})
		entry.Debugf(format, args...)
	}
}

// Infof logs a message at level Info.
func Infof(format string, args ...interface{}) {
	if logger.Level >= logrus.InfoLevel {
		// Create a log entry with empty fields
		entry := logger.WithFields(logrus.Fields{})
		entry.Infof(format, args...)
	}
}

// Warnf logs a message at level Warn.
func Warnf(format string, args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		// Create a log entry with empty fields
		entry := logger.WithFields(logrus.Fields{})
		entry.Warnf(format, args...)
	}
}

// Errorf logs a message at level Error.
func Errorf(format string, args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		// Create a log entry with empty fields
		entry := logger.WithFields(logrus.Fields{})
		entry.Errorf(format, args...)
	}
}

// Fatalf logs a message at level Fatal and exits the program.
func Fatalf(format string, args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		// Create a log entry with empty fields
		entry := logger.WithFields(logrus.Fields{})
		entry.Fatalf(format, args...)
	}
}

// contextKey type for context values to avoid collisions.
type contextKey string

// RequestIDContextKey is the key used to store request_id in context.Context.
// Middleware should set it via context.WithValue(c.Request.Context(), logger.RequestIDContextKey, requestID).
const RequestIDContextKey contextKey = "request_id"

// Formatter implements logrus.Formatter
type formatter struct {
	prefix string
}

// Format builds the log message according to custom format.
// If entry has "request_id" in Data, prepends [request_id] to the message.
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var sb bytes.Buffer

	sb.WriteString(strings.ToUpper(entry.Level.String()))
	sb.WriteString(" ")
	sb.WriteString(entry.Time.Format(time.RFC3339))
	sb.WriteString(" ")
	sb.WriteString(f.prefix)
	if rid, ok := entry.Data["request_id"]; ok {
		if ridStr, ok := rid.(string); ok && ridStr != "" {
			sb.WriteString("[")
			sb.WriteString(ridStr)
			sb.WriteString("] ")
		}
	}
	sb.WriteString(entry.Message)
	sb.WriteByte('\n')

	return sb.Bytes(), nil
}

// WithRequestID returns a log entry that includes request_id in every log line.
func WithRequestID(requestID string) *logrus.Entry {
	return logger.WithField("request_id", requestID)
}

// FromContext returns a log entry with request_id from ctx if present; otherwise without request_id.
func FromContext(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logger.WithFields(logrus.Fields{})
	}
	if rid, ok := ctx.Value(RequestIDContextKey).(string); ok && rid != "" {
		return logger.WithField("request_id", rid)
	}
	return logger.WithFields(logrus.Fields{})
}

// LogStart logs START and returns ctx and start time. Use at handler/method entry; call LogFinish before every return.
// In controllers pass c.Request.Context(); in services pass the ctx received from the controller.
func LogStart(ctx context.Context, spanName string) (context.Context, time.Time) {
	start := time.Now()
	FromContext(ctx).Infof("START %s", spanName)
	return ctx, start
}

// LogFinish logs FINISH with SUCCESS or FAIL and duration. Call before every return in handlers and service methods.
func LogFinish(ctx context.Context, spanName string, err error, start time.Time) {
	durMs := time.Since(start).Seconds() * 1000
	status := "SUCCESS"
	if err != nil {
		status = "FAIL"
	}
	FromContext(ctx).Infof("FINISH %s (%s) duration=%.2fms", spanName, status, durMs)
}
