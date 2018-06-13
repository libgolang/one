package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/magiconair/properties"
)

var (
	globalLoadedConfig = false
	globalLevels       map[string]Level
	globalLoggers      = map[string]Logger{}
	globalWriters      = []Writer{getDefaultWriter()}
	globalLogger       = New("")
	globalTraceEnabled = false
)

// Logger interface exposed to users
type Logger interface {
	Error(format string, args ...interface{}) error
	Warn(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Panic(format string, args ...interface{})
	SetLevel(Level)
}

// Writer writer interface
type Writer interface {
	WriteLog(name string, logLevel Level, format string, args []interface{})
	SetLevel(level Level)
}

// logger instance
type logger struct {
	level Level
	name  string
}

// New constructor
func New(name string) Logger {
	// level
	var lvl Level
	if l, ok := globalLevels[name]; ok {
		lvl = l
	} else {
		lvl = globalLevels[""]
	}

	l := &logger{lvl, name}
	globalLoggers[name] = l

	if !globalLoadedConfig {
		LoadLogProperties()
		globalLoadedConfig = true
	}

	return l
}

// SetDefaultLevel sets the default logging level. It defaults to WARN
func SetDefaultLevel(l Level) {
	globalLevels[""] = l
	globalLogger.SetLevel(l)
}

// SetTrace when set to true, the log will print file names and line numbers
func SetTrace(trace bool) {
	globalTraceEnabled = trace
}

// IsTraceEnabled whether printing of file names and line numbers is enabled
func IsTraceEnabled() bool {
	return globalTraceEnabled
}

// SetLoggerLevels sets the levels for all existing loggers and future loggers
func SetLoggerLevels(levels map[string]Level) {
	//
	for k, lev := range levels {
		if log, ok := globalLoggers[k]; ok {
			log.SetLevel(lev)
		}
	}

	// make sure there is always a root logger level
	if _, ok := levels[""]; !ok {
		levels[""] = WARN
	}

	//
	globalLevels = levels
}

// SetWriters sets the writers for all the loggers
func SetWriters(w []Writer) {
	globalWriters = w
}

// Error logs at error level
func (l *logger) Error(format string, a ...interface{}) error {
	PrintLog(l.name, l.level, ERROR, format, a)
	return fmt.Errorf(format, a...)
}

// Info logs at info level
func (l *logger) Info(format string, a ...interface{}) {
	PrintLog(l.name, l.level, INFO, format, a)
}

// Warn logs at wanr level
func (l *logger) Warn(format string, a ...interface{}) {
	PrintLog(l.name, l.level, WARN, format, a)
}

// Debug logs at debug level
func (l *logger) Debug(format string, a ...interface{}) {
	PrintLog(l.name, l.level, DEBUG, format, a)
}

// Panic error and exit
func (l *logger) Panic(format string, a ...interface{}) {
	PrintLog(l.name, l.level, ERROR, format, a)
	panic("panic!")
}

// SetLevel set logger level
func (l *logger) SetLevel(level Level) {
	l.level = level
}

// PrintLog sends a log message to the writers.
// name: logger name
// loggerLevel: the level of the logger implementation
// logLevel: the level of the message. If the level of the message is greater than loggerLevel the log will bi discarted
// format: log format.  See fmt.Printf
// a...: arguments.  See fmt.Printf
func PrintLog(name string, loggerLevel, logLevel Level, format string, a []interface{}) {
	if loggerLevel < logLevel {
		return
	}
	for _, w := range globalWriters {
		w.WriteLog(name, logLevel, format, a)
	}
}

// Error log to root logger
func Error(format string, a ...interface{}) error {
	return globalLogger.Error(format, a...)
}

// Info log to root logger
func Info(format string, a ...interface{}) {
	globalLogger.Info(format, a...)
}

// Warn log to root logger
func Warn(format string, a ...interface{}) {
	globalLogger.Warn(format, a...)
}

// Debug log to root logger
func Debug(format string, a ...interface{}) {
	globalLogger.Debug(format, a...)
}

// Panic log to root logger
func Panic(format string, a ...interface{}) {
	globalLogger.Panic(format, a...)
}

// resolves configuration
func getDefaultLevel(def Level) Level {
	str, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return def
	}
	l := strToLevel(str)
	if l == OTHER {
		return def
	}
	return l
}

// resolves configuration
func getDefaultWriter() Writer {
	return &WriterStdout{WARN}
}

// LoadLogProperties loads properties from configuration file in LOG_CONFIG
func LoadLogProperties() {
	cfgFile, ok := os.LookupEnv("LOG_CONFIG")
	if !ok {
		return
	}

	props, err := properties.LoadFile(cfgFile, properties.UTF8)
	if err != nil {
		return
	}

	//
	// Trace
	//
	if props.GetString("log.trace", "false") == "true" {
		SetTrace(true)
	}

	//
	// Levels
	//
	logLevels := make(map[string]Level)
	logLevels[""] = strToLevel(props.GetString("log.level", ""))
	for k, v := range props.Map() {
		if !strings.HasPrefix(k, "log.level.") {
			continue
		}

		parts := strings.Split(k, ".")
		if len(parts) != 3 {
			continue
		}
		loggerName := parts[2] //log.level.name1=stdout|file
		logLevels[loggerName] = strToLevel(v)
	}

	//
	// Writers
	//
	logWriters := make([]Writer, 0)
	processed := make(map[string]bool)
	for k := range props.Map() {
		if strings.HasPrefix(k, "log.writer.") {
			parts := strings.Split(k, ".")
			if len(parts) != 4 {
				continue
			}
			loggerName := parts[2]

			// already set
			if _, ok := processed[loggerName]; ok {
				continue
			}

			var writer Writer
			loggerType := props.GetString(fmt.Sprintf("log.writer.%s.type", loggerName), "stdout")
			loggerLevel := strToLevel(props.GetString(fmt.Sprintf("log.writer.%s.level", loggerName), "DEBUG"))
			if loggerType == "stdout" {
				writer = &WriterStdout{level: loggerLevel}
			} else if loggerType == "file" {
				size := props.GetInt64(fmt.Sprintf("log.writer.%s.maxSize", loggerName), int64(Gigabyte))
				maxfiles := props.GetInt(fmt.Sprintf("log.writer.%s.maxFiles", loggerName), 10)
				dir := props.GetString(fmt.Sprintf("log.writer.%s.dir", loggerName), "./log")
				name := props.GetString(fmt.Sprintf("log.writer.%s.name", loggerName), loggerName)
				writer = NewFileWriter(dir, name, FileSize(size), maxfiles, loggerLevel)
			} else {
				continue
			}
			processed[loggerName] = true
			logWriters = append(logWriters, writer)
		}
	}

	SetLoggerLevels(logLevels)
	if len(logWriters) > 0 {
		SetWriters(logWriters)
	}
}
