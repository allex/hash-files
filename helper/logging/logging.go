package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

const (
	levelError = iota
	levelWarn
	levelInfo
	levelDebug
)

type logger struct {
	out       io.Writer
	prefix    string
	isDiscard int32 // atomic boolean: whether out == io.Discard
}

func (l *logger) Print(msg string) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	fmt.Fprint(l.out, msg)
}

func newLogger(out io.Writer, prefix string) *logger {
	return &logger{out: out, prefix: prefix}
}

var (
	errorLogger = newLogger(os.Stderr, red+"[ERROR]\t")
	warnLogger  = newLogger(os.Stderr, yellow+"[WARN]\t")
	infoLogger  = newLogger(os.Stderr, green+"[INFO]\t")
	debugLogger = newLogger(os.Stderr, magenta+"[DEBUG]\t")
	logLevel    = levelInfo
	mu          sync.Mutex
)

// SetLogLevel sets the current level of logging output.
// The available options are: debug, info, warn, error
func SetLogLevel(level string) error {
	mu.Lock()
	defer mu.Unlock()

	mapping := map[string]int{
		"error": levelError,
		"warn":  levelWarn,
		"info":  levelInfo,
		"debug": levelDebug,
	}

	if val, ok := mapping[strings.ToLower(level)]; !ok {
		return errors.New("invalid log-level string")
	} else {
		logLevel = val
	}

	return nil
}

func Stderr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func Stdout(format string, a ...any) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func Error(message string) {
	if logLevel >= levelError {
		errorLogger.Print(message + reset)
	}
}

func Warn(message string) {
	if logLevel >= levelWarn {
		warnLogger.Print(message + reset)
	}
}

func Info(message string) {
	if logLevel >= levelInfo {
		infoLogger.Print(message + reset)
	}
}

func Debug(message string) {
	if logLevel >= levelDebug {
		debugLogger.Print(message + reset)
	}
}
