package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	projectName = "smart-money-bot"
)

// Define log level string
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warning"
	ErrorLevel = "error"

	fieldCaller = "caller"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	cyan    = 36
	gray    = 37
)

const (
	maximumCallerDepth int = 25
	knownLogFrames     int = 4
)

var (
	minimumCallerDepth = 1
	// Used for caller information initialisation
	callerInitOnce sync.Once

	// qualified package name, cached at first use
	logPackage string

	// Levels contains all valid log levels
	Levels = []string{DebugLevel, InfoLevel, WarnLevel, ErrorLevel}

	loggers      Loggers
	currentLevel string
)

func colored(level logrus.Level, line string) string {
	var levelColor int
	switch level {
	case logrus.DebugLevel:
		levelColor = cyan
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	return fmt.Sprintf("\033[1;%dm%s\033[0m", levelColor, line)
}

func isTerminal() bool {
	return terminal.IsTerminal(int(os.Stderr.Fd()))
}

// CurrentLevel return current log level
func CurrentLevel() string {
	return currentLevel
}

// InitLogger initialize loggers
func InitLogger(logLevel, logPath, errorLogLevel, errorLogPath string) (err error) {
	logger, err := newFileLogger(logLevel, logPath)
	if err != nil {
		return
	}
	errLogger, err := newFileLogger(errorLogLevel, errorLogPath)
	if err != nil {
		return
	}
	loggers := []*logrus.Logger{logger, errLogger}
	if isTerminal() {
		consoleLogger, err := newConsoleLogger(logLevel)
		if err != nil {
			return err
		}
		loggers = append(loggers, consoleLogger)
	}
	SetLoggers(loggers...)
	currentLevel = logLevel
	Infof("Init logger at %s, error log at %s", logPath, errorLogPath)
	return
}

// InitDevopsLogger initialize logger for devops
func InitDevopsLogger(logLevel string) {
	consoleLogger, err := newConsoleLogger(logLevel)
	if err != nil {
		panic(err)
	}
	loggers = []*logrus.Logger{consoleLogger}
	SetLoggers(loggers...)
	currentLevel = logLevel
	Info("Init dev ops logger")
}

// InitTestLogger initialize logger for unit test
func InitTestLogger() {
	consoleLogger, err := newConsoleLogger(DebugLevel)
	if err != nil {
		panic(err)
	}
	loggers = []*logrus.Logger{consoleLogger}
	SetLoggers(loggers...)
	currentLevel = DebugLevel
	Info("Init test logger")
}

func newFileLogger(logLevel, logPath string) (logger *logrus.Logger, err error) {
	var level logrus.Level
	if level, err = logrus.ParseLevel(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level: %s", logLevel)
		return
	}
	logger = logrus.New()
	logger.Formatter = &textFormatter{colored: false}
	logger.Level = level
	fd, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open logfile %s: %s\n", logPath, err)
		return
	}
	logger.Out = fd
	return
}

func newConsoleLogger(logLevel string) (logger *logrus.Logger, err error) {
	var level logrus.Level
	if level, err = logrus.ParseLevel(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level: %s", logLevel)
		return
	}
	logger = logrus.New()
	logger.Formatter = &textFormatter{colored: false}
	logger.Level = level
	logger.Out = os.Stderr
	return
}

// InitConsoleLogger initialize logger for console only
func InitConsoleLogger(logLevel string) (err error) {
	consoleLogger, err := newConsoleLogger(logLevel)
	if err != nil {
		return err
	}
	SetLoggers(consoleLogger)
	return
}

type textFormatter struct {
	colored bool
}

// Format format a log line
func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	caller := entry.Data[fieldCaller]

	b := new(bytes.Buffer)
	f.appendKeyValue(b, "time", entry.Time.Format(time.RFC3339), false)
	f.appendKeyValue(b, "level", entry.Level.String(), false)
	f.appendKeyValue(b, fieldCaller, caller, true)
	for key, value := range entry.Data {
		if key == fieldCaller {
			continue
		}
		f.appendKeyValue(b, key, value, false)
	}
	f.appendKeyValue(b, "msg", entry.Message, true)
	b.WriteByte('\n')

	if f.colored {
		line := b.String()
		b = new(bytes.Buffer)
		fmt.Fprint(b, colored(entry.Level, line))
	}

	return b.Bytes(), nil
}

func (f *textFormatter) appendKeyValue(
	b *bytes.Buffer, key string, value interface{}, quoted bool,
) {

	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if !quoted {
			fmt.Fprintf(b, "%s", value)
		} else {
			// we don't use go's quoted string here, because out feilds like 'msg' may contains
			// new line characters
			fmt.Fprintf(b, "\"%s\"", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

// getPackageName reduces a fully qualified function name to the packagname
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

// getCaller retrieves the name of the first non-log package calling function
func getCaller() string {
	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	// cache this package's fully-qualified name
	callerInitOnce.Do(
		func() {
			logPackage = getPackageName(runtime.FuncForPC(pcs[0]).Name())

			// now that we have the cache, we can skip a minimum count of known-log functions
			// XXX this is dubious, the number of frames may vary store an entry in a logger interface
			minimumCallerDepth = knownLogFrames
		},
	)

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logPackage {
			parts := strings.Split(f.File, projectName)
			file := parts[len(parts)-1]
			file = strings.TrimPrefix(file, "/")
			return fmt.Sprintf("%s:%d %s()", file, f.Line, f.Function)
		}
	}

	// if we got here, we failed to find the caller's context
	return "???"
}

func logEntry(logger *logrus.Logger) *logrus.Entry {
	return logger.WithField(fieldCaller, getCaller())
}

// SetLoggers set the standard logger
func SetLoggers(log ...*logrus.Logger) {
	oldLoggers := loggers
	loggers = Loggers(log)

	for _, logger := range oldLoggers {
		f, ok := logger.Out.(*os.File)
		if !ok {
			continue
		}
		if f == os.Stdout || f == os.Stderr {
			continue
		}
		if err := f.Close(); err != nil {
			Warnf("failed to close log file, err: %s", err)
			continue
		}
	}
	return
}

// Logger returns current logger
func Logger() Loggers {
	return loggers
}

// Loggers for logger list
type Loggers []*logrus.Logger

// AddHook add hook for logrus
func (loggers Loggers) AddHook(hook logrus.Hook) {
	for _, logger := range loggers {
		logger.AddHook(hook)
	}
}

// Debug logs a message at level Debug on loggers
func (loggers Loggers) Debug(args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Debug(args...)
	}
}

// Info logs a message at level Info on loggers
func (loggers Loggers) Info(args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Info(args...)
	}
}

// Warn logs a message at level Warn on loggers
func (loggers Loggers) Warn(args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Warn(args...)
	}
}

// Error logs a message at level Error on loggers
func (loggers Loggers) Error(args ...interface{}) {
	fmtError := FormatErrorStack(fmt.Sprint(args...))
	for _, logger := range loggers {
		logEntry(logger).Error(fmtError)
	}
}

// Debugf logs a message at level Debug on loggers
func (loggers Loggers) Debugf(format string, args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Debugf(format, args...)
	}
}

// Infof logs a message at level Info on loggers
func (loggers Loggers) Infof(format string, args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Infof(format, args...)
	}
}

// Warnf logs a message at level Warn on loggers
func (loggers Loggers) Warnf(format string, args ...interface{}) {
	for _, logger := range loggers {
		logEntry(logger).Warnf(format, args...)
	}
}

// Errorf logs a message at level Error on loggers
func (loggers Loggers) Errorf(format string, args ...interface{}) {
	fmtError := FormatErrorStack(format)
	for _, logger := range loggers {
		logEntry(logger).Errorf(fmtError, args...)
	}
}

// AddHook add hook for logrus
func AddHook(hook logrus.Hook) {
	loggers.AddHook(hook)
}

// Debug logs a message at level Debug on loggers
func Debug(args ...interface{}) {
	loggers.Debug(args...)
}

// Info logs a message at level Info on loggers
func Info(args ...interface{}) {
	loggers.Info(args...)
}

// Warn logs a message at level Warn on loggers
func Warn(args ...interface{}) {
	loggers.Warn(args...)
}

// Error logs a message at level Error on loggers
func Error(args ...interface{}) {
	loggers.Error(args...)
}

// Debugf logs a message at level Debug on loggers
func Debugf(format string, args ...interface{}) {
	loggers.Debugf(format, args...)
}

// Infof logs a message at level Info on loggers
func Infof(format string, args ...interface{}) {
	loggers.Infof(format, args...)
}

// Warnf logs a message at level Warn on loggers
func Warnf(format string, args ...interface{}) {
	loggers.Warnf(format, args...)
}

// Errorf logs a message at level Error on loggers
func Errorf(format string, args ...interface{}) {
	loggers.Errorf(format, args...)
}

// FormatErrorStack  Remove  project name  attached by errors.ErrorStack,
func FormatErrorStack(format string) string {
	if lines := strings.Split(format, "\n"); len(lines) > 0 {
		for i := len(lines); i > 0; i-- {
			parts := strings.Split(lines[i-1], fmt.Sprintf("%s/", projectName))
			lines[i-1] = parts[len(parts)-1]
		}
		return strings.Join(lines, "\n")
	}
	return format
}
