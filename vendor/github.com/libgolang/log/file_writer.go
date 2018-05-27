package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

// FileSize represents the size of a file
type FileSize int64

const (
	_ = iota //

	// Kilobyte usage (Kilobyte * 2) = 2Kb
	Kilobyte FileSize = 1 << (10 * iota) // 1 << (10*1)

	// Megabyte usage (Megabyte * 2) = 2Mb
	Megabyte // 1 << (10*2)

	// Gigabyte usage (Gigabyte * 2) = 2Gb
	Gigabyte // 1 << (10*3)

	// Terabyte usage (Terabyte * 2) = 2Tb
	Terabyte // 1 << (10*4)

	// Petabyte usage (Petabyte * 2) = 2Pb
	Petabyte // 1 << (10*5)

	// Exabyte usage (Exabyte * 2) = 2Pb
	Exabyte // 1 << (10*6)
)

type fileWriter struct {
	level           Level
	logDir          string
	fileNamePattern string
	file            *os.File
	fileLen         int64
	maxSize         FileSize
	numFiles        int
	logQueue        chan *writeLogMsg
}

// NewFileWriter retuns a new instance of fileWriter.  Parameter logdir is the
// directory where logs will be written.  Parameter fileName is the name to use
// when creating the files.  Param maxSize is the maximum size of a log, when the size
// is exceded, the log is rotated.  maxNumFiles is the maximum number of logs to keep.
func NewFileWriter(logDir, fileName string, maxSize FileSize, maxNumFiles int, level Level) Writer {
	it := &fileWriter{}

	it.logDir = logDir
	it.fileNamePattern = fileName
	it.maxSize = maxSize
	it.numFiles = maxNumFiles
	it.logQueue = make(chan *writeLogMsg, 100)
	it.level = level

	it.init()

	return it
}

type writeLogMsg struct {
	name   string
	mLevel Level
	format string
	args   []interface{}
}

// WriteLog implementation of logger.Writer
func (fw *fileWriter) WriteLog(
	name string,
	mLevel Level,
	format string,
	args []interface{},
) {
	if fw.level < mLevel {
		return
	}
	fw.logQueue <- &writeLogMsg{name, mLevel, format, args}
}

func (fw *fileWriter) SetLevel(level Level) {
	fw.level = level

}

func (fw *fileWriter) writeLog(m *writeLogMsg) {
	var preFormat string
	if IsTraceEnabled() {
		_, file, line, _ := runtime.Caller(4)
		preFormat = fmt.Sprintf("%s %s [%s] %s:%d %s\n", time.Now().Format(time.RFC3339), m.mLevel, m.name, file, line, m.format)
	} else {
		preFormat = fmt.Sprintf("%s %s [%s] %s\n", time.Now().Format(time.RFC3339), m.mLevel, m.name, m.format)
	}

	str := fmt.Sprintf(preFormat, m.args...)
	fw.fileLen += int64(len(str))
	fmt.Fprint(fw.file, str)

	if fw.fileLen > int64(fw.maxSize) {
		// todo: rotate file
		fw.rotateFile()
	}
}

func (fw *fileWriter) init() {
	// Init file
	fw.rotateFile()

	// start the log writer queue
	go func() {
		for m := range fw.logQueue {
			fw.writeLog(m)
		}
	}()
}

func (fw *fileWriter) rotateFile() {

	// If not open, then create/open file
	if fw.file == nil {
		file, err := os.OpenFile(fw.getFileName(0), os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(err)
		}
		fw.file = file
		fw.rotateFile()
	} else {
		nfo, err := fw.file.Stat()
		if err != nil {
			panic(err)
		}
		if nfo.Size() < int64(fw.maxSize) {
			return // bail out, file not big enough
		}

		// rename all files file-n.log to file-(n+1).log
		for i := fw.numFiles; i >= 0; i-- {
			oldName := fw.getFileName(i)
			if _, err := os.Stat(oldName); os.IsNotExist(err) {
				continue
			}

			newName := fw.getFileName(i + 1)
			if _, err := os.Stat(newName); os.IsNotExist(err) {
				_ = os.Rename(oldName, newName)
			} else {
				_ = os.Remove(newName)
			}
		}

		// remove last file
		lastFile := fw.getFileName(fw.numFiles + 1)
		if _, err := os.Stat(lastFile); !os.IsNotExist(err) {
			_ = os.Remove(lastFile)
		}

		// close and re-open with new name
		_ = fw.file.Close()
		file, err := os.OpenFile(fw.getFileName(0), os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(err)
		}
		fw.file = file
		fw.fileLen = 0
	}

}

func (fw *fileWriter) getFileName(r int) string {
	if _, err := os.Stat(fw.logDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fw.logDir, 0755); err != nil {
			panic(err)
		}
	}
	return path.Join(fw.logDir, fmt.Sprintf("%s-%d.log", fw.fileNamePattern, r))
}
