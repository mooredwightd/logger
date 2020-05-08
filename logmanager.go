// Log Manager (LM)
//
// The is a service that provides the ability to manage a set of logs, based on Log type,
// and writes message to the log.  When creating a LM, one specifies the log file struct
// created by one of the existing types or a custom type.  A custom log file must implement
// the logger.LogWriter interface.
//
// Base implementation provides
// File: static log file with no rotation or file size limit
// SizeLimitedFile: log file that rotates to a new volume when it reaches a specified size.
// DailyFile: log file that rotates at midnight each day
// TimedFile: log fiel that rotates at a specified interval.
//
// Example:
//      // Create a log file via a model (daily, schedule, size limited, etc.)
// 	f, err := logger.DailyFile(name)
//      if err != nil {
//          panic("Error creating log file.")
//      }
//      // Instantiate a log manager
//      l = logger.LogManger("MyApp", f)
//      // Log an event (message)
// 	l.Debug("NEWCLIENT", "Created new client", map[string]string{"app":"Myapp", "logfile":name})
//
//      f2, err := logger.SizeLimitedFile("/somepath/logs/sizedfile", 2 * logger.Mbyte)
//      l.AddLogger(f2)

package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Private constants
const ()

type LogWriter interface {
	io.WriteCloser
}

type Logger interface {
	New(app string, lwc LogWriter) *Log
	Close()
	AddLogger(lwc LogWriter)
	LogEvent(sev Severity, msgId string, msg string, params map[string]string)
}

type Log struct {
	version    string
	hostname   string
	appname    string
	filter     Severity
	logModules []LogWriter
	formatter  EventFormatter
}

type EventMsg struct {
	Timestamp time.Time         `json:"timestamp"`
	Sev       string            `json:"severity"`
	Hostname  string            `json:"hostname"`
	Appname   string            `json:"appname"`
	Pid       int               `json:"pid"`
	MsgId     string            `json:"msg_id"`
	Msg       string            `json:"message"`
	Params    map[string]string `json:"params"`
}

var (
	// Invalid argument error
	InvalidArgumentError error = errors.New("Invalid Argument Exception")
)

func init() {
}

// Create a new LogManager.
// app is a string distinguishing the application in the logs
// lwc is a LogWriterClose which receives the logged messages.
func LogManger(app string, lwc LogWriter) *Log {
	h, _ := os.Hostname()
	l := &Log{hostname: h, appname: app}
	l.logModules = make([]LogWriter, 1)
	l.logModules[0] = lwc
	l.SetFormatter(Json())
	l.filter = Debug
	return l
}

// Set the event formatter for the log record
// Parameter ef must implement the logger.EventFormatter interface.
func (l *Log) SetFormatter(ef EventFormatter) {
	l.formatter = ef
}

// Set the event message filter level.
// The filter only writes for at a Severity level >= the current filter.
// If the Severity value is invalid, and error is returned.
func (l *Log) SetFilter(sev Severity) (err error) {
	if l.filter < SeverityMinLevel || l.filter > SeverityMaxLevel {
		return InvalidArgumentError
	}
	l.filter = sev
	return err
}

// Returns the current filter level
func (l *Log) GetFilter() Severity {
	return l.filter
}

// Add another logger to the manager
// lwc is a LogWriterCloser
func (l *Log) AddLogger(lwc LogWriter) {
	l.logModules = append(l.logModules, lwc)
}

// Close all log interfaces
func (l *Log) Close() {
	for _, mod := range l.logModules {
		mod.Close()
	}
	l.logModules = nil
}

// Write a message to the log(s)
func (l *Log) LogEvent(sev Severity, msgId string, msg string, params map[string]string) {
	if sev > l.filter {
		return
	}

	em := validateEventMsg(l.newEventMsg(sev, msgId, msg, params))
	str, err := l.formatter.Format(*em)
	if err != nil {
		log.Println("logger.LogEvent WARN: Error in formatting message. No log output generated.")
		return
	}
	bMsg := []byte(str)
	for _, mod := range l.logModules {
		mod.Write(bMsg)
	}
}

// Convenience fnction to log an EMERGENCY level message
// Applicability: System is unusable
func (l *Log) Emergency(msgId string, msg string, params map[string]string) {
	l.LogEvent(Emergency, msgId, msg, params)
}

// Convenience fnction to log an ALERT level message
// Applicability: Action must be taken immediately.
// 	Example: Entire website down, database unavailable, etc. This should trigger the
// 	SMS alerts and wake you up.
func (l *Log) Alert(msgId string, msg string, params map[string]string) {
	l.LogEvent(Alert, msgId, msg, params)
}

// Convenience fnction to log a CRITCALL level message
// Critical conditions
// Applicability: Application component unavailable, unexpected exception.
func (l *Log) Critical(msgId string, msg string, params map[string]string) {
	l.LogEvent(Critical, msgId, msg, params)
}

// Convenience fnction to log an ERROR level message
// Applicability: Runtime errors that do not require immediate action but should typically
// 	be logged and monitored
func (l *Log) Error(msgId string, msg string, params map[string]string) {
	l.LogEvent(Error, msgId, msg, params)
}

// Convenience fnction to log a WARNING level message
// Applicability: Exceptional occurrences that are not errors.
// 	Example: Use of deprecated APIs, poor use of an API, undesirable things
// 	that are not necessarily wrong.
func (l *Log) Warning(msgId string, msg string, params map[string]string) {
	l.LogEvent(Warning, msgId, msg, params)
}

// Convenience fnction to log a NOTICE level message
// Applicability: Normal but significant events.
func (l *Log) Notice(msgId string, msg string, params map[string]string) {
	l.LogEvent(Notice, msgId, msg, params)
}

// Convenience fnction to log an INFO level message
// Applicability: Detailed debug information.
func (l *Log) Info(msgId string, msg string, params map[string]string) {
	l.LogEvent(Info, msgId, msg, params)
}

// Convenience fnction to log a DEBUG level message.
// Applicability: Detailed debug information.
func (l *Log) Debug(msgId string, msg string, params map[string]string) {
	l.LogEvent(Debug, msgId, msg, params)
}

func (l *Log) newEventMsg(sev Severity, msgId string, msg string, params map[string]string) *EventMsg {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("Error writing log: %s\n", x)
			return
		}
	}()

	em := EventMsg{
		Sev:       sev.String(),
		Pid:       os.Getpid(),
		Hostname:  l.hostname,
		Appname:   l.appname,
		MsgId:     msgId,
		Timestamp: time.Now().Round(time.Microsecond),
		Params:    params,
		Msg:       msg}

	return &em
}

// Get the caller function/method name in the stack.
// Returns a string of the function name.
func GetCaller() string {
	_, s, _ := getStack(0)
	return s[3] // Skip runtime.Callers, getStack, GetCaller frames
}

func getStack(skip int) (int, []string, error) {
	var pc []uintptr
	var f runtime.Frame

	pc = make([]uintptr, 100)
	n := runtime.Callers(skip, pc)
	s := make([]string, n)
	fs := runtime.CallersFrames(pc)

	f, ok := fs.Next()
	for i := 0; i < n; i++ {
		s[i] = fmt.Sprintf("%s, %s, Line %d",
			f.File[strings.LastIndex(f.File, "/"):], f.Function, f.Line)
		f, ok = fs.Next()
		if !ok {
			break
		}
	}
	return n, s, nil
}
