package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production
	DebugLevel LogLevel = iota
	// InfoLevel is the default logging priority
	InfoLevel
	// WarningLevel logs are more important than Info, but don't need individual human review
	WarningLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs
	ErrorLevel
)

var (
	levelNames = map[LogLevel]string{
		DebugLevel:   "DEBUG",
		InfoLevel:    "INFO",
		WarningLevel: "WARN",
		ErrorLevel:   "ERROR",
	}

	levelColors = map[LogLevel]string{
		DebugLevel:   "\033[36m", // Cyan
		InfoLevel:    "\033[32m", // Green
		WarningLevel: "\033[33m", // Yellow
		ErrorLevel:   "\033[31m", // Red
	}
)

// Logger is a leveled logger
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	logger   *log.Logger
	out      io.Writer
	useColor bool
}

// Global logger instance
var std = &Logger{
	level:    WarningLevel, // Default to Warning as requested
	out:      os.Stderr,
	useColor: true,
}

// SetLevel sets the global log level
func SetLevel(level LogLevel) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.level
}

// SetOutput sets the output destination
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
	std.logger = log.New(w, "", 0)
}

// SetColor enables or disables colored output
func SetColor(enabled bool) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.useColor = enabled
}

// Debug logs a message at DebugLevel
func Debug(format string, args ...interface{}) {
	std.log(DebugLevel, format, args...)
}

// DebugJSON logs a JSON object at DebugLevel with pretty formatting
func DebugJSON(name string, obj interface{}) {
	std.logJSON(DebugLevel, name, obj)
}

// Info logs a message at InfoLevel
func Info(format string, args ...interface{}) {
	std.log(InfoLevel, format, args...)
}

// InfoJSON logs a JSON object at InfoLevel with pretty formatting
func InfoJSON(name string, obj interface{}) {
	std.logJSON(InfoLevel, name, obj)
}

// Warning logs a message at WarningLevel
func Warning(format string, args ...interface{}) {
	std.log(WarningLevel, format, args...)
}

// Error logs a message at ErrorLevel
func Error(format string, args ...interface{}) {
	std.log(ErrorLevel, format, args...)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	levelName := levelNames[level]

	if l.useColor {
		color := levelColors[level]
		reset := "\033[0m"
		log.SetOutput(l.out)
		log.Printf("%s[%-5s]%s %s\n", color, levelName, reset, msg)
	} else {
		log.SetOutput(l.out)
		log.Printf("[%-5s] %s\n", levelName, msg)
	}
}

// logJSON logs a JSON object with pretty formatting
func (l *Logger) logJSON(level LogLevel, name string, obj interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		l.log(level, "%s: <failed to marshal: %v>", name, err)
		return
	}

	levelName := levelNames[level]

	if l.useColor {
		color := levelColors[level]
		reset := "\033[0m"
		log.SetOutput(l.out)
		log.Printf("%s[%-5s]%s %s:\n%s%s\n", color, levelName, reset, name, string(data), reset)
	} else {
		log.SetOutput(l.out)
		log.Printf("[%-5s] %s:\n%s\n", levelName, name, string(data))
	}
}

// ParseLevel parses a string level name to LogLevel
func ParseLevel(level string) (LogLevel, error) {
	switch level {
	case "DEBUG":
		return DebugLevel, nil
	case "INFO":
		return InfoLevel, nil
	case "WARN", "WARNING":
		return WarningLevel, nil
	case "ERROR":
		return ErrorLevel, nil
	default:
		return WarningLevel, fmt.Errorf("unknown log level: %s", level)
	}
}
