package log

import (
	"fmt"
	"runtime"
	"time"
)

// WriterStdout writes to the standard output
type WriterStdout struct {
	level Level
}

// WriteLog implementation of logger.Writer
func (w *WriterStdout) WriteLog(
	name string,
	mLevel Level,
	format string,
	args []interface{},
) {
	if w.level < mLevel {
		return
	}

	var preFormat string
	if IsTraceEnabled() {
		_, file, line, _ := runtime.Caller(4)
		preFormat = fmt.Sprintf("%s %s [%s] %s:%d %s\n", time.Now().Format(time.RFC3339), mLevel, name, file, line, format)
	} else {
		preFormat = fmt.Sprintf("%s %s [%s] %s\n", time.Now().Format(time.RFC3339), mLevel, name, format)
	}

	fmt.Printf(preFormat, args...)
}

// SetLevel sets the writer level
func (w *WriterStdout) SetLevel(level Level) {
	w.level = level
}

/*
type writerChannelMsg struct {
	name   string
	mLevel Level
	format string
	args   []interface{}
}
*/
