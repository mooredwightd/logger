// File Logger (FL)
// A file logger writes to a log "file", and implementes the logger.FileLogger interface.
// These are used, and registered with the LogManager for logging. Any repository/storage type ould be
// implemented using the interface. A policy type is also used. For custom policies, a PolicyNone can be used.
// The pre-defined file loggers are
//     File (static, non-limited file)
//     SizeLimitedFile, which rotates once the file size limit is reached.
//     DailyFile, which rotates each day at 00:00.00.
//     TimedFile, which rotates at a defined interface, e.g. every 2 hours.
//
package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// File size constants
	Kbyte int64 = 1024
	Mbyte int64 = 1000 * Kbyte
	Gbyte int64 = 1000 * Mbyte

	// Used to indicate a file does not have a file size constraint
	LogSizeNotLimited = -1
	LogMinFileSize = Mbyte
	LogMaxFileSize = (500 * Mbyte)


	// path/prefix"."date_and-or_volume"."log.
	logFilenameExtension string = "log"
	logFilenameDailyFormat string = "2016-01-01"
	logFilenameTimeFormat string = "2016-01-01T23.01.01"
	logFilenameVolumeFormat string = "%04.4d"

	// Volumes from 1 to 9999
	logMaxVolNumber int = 9999

	// Open mode is u=rw, g=rw, o=none
	logDefaultFileMode os.FileMode = 0660
	logDefaultOpenFlags int = os.O_CREATE | os.O_APPEND

	// Indicates the low water mark to cause a file rotation.
	logHighWaterMark = (2*Kbyte)


)

type FileWriter interface {
	LogRotateCheck() bool
	LogRotate() bool
	LogPolicy() PolicyType
	LogFilename() string
}

// Implements the a log file with a name (prefix), policy, and duration for file rotation policy.
type LogFile struct {
				    // prefix : Prefix for the filename. This includes the path.
	prefix        string
				    // currentFile: The current file for writing.
	currentFile   string
				    // 	The type of policy determines the remaining part.
	policy        PolicyType
	volNo         int16         // Used for static files or PolicyFileSzie
	fileSizeLimit int64         // Use for PolicyFileSize
				    // The current io.Writer for this log.
	f             io.WriteCloser
	cycle         time.Duration // Time rotation cycle
				    // Used to create a timer event for log rotation e.g. Daily, Scheduled
	ltimer        *LogTimer
	filenameGen   func() string
	rotateCheck   func() bool
	rotate        func() bool
	newTimer      func() *LogTimer
	sync.Mutex
}

// Public methods
//


// Creates a simple, non-rotating log file. Two File logs with the same name (prefix) point to the
// same file. The name parameter is a full file path and filename, with no extension.
//
// If an error occurs, returns nil, and an error.
func File(name string) (lf *LogFile, err error) {
	lf = &LogFile{
		prefix: name,
		policy: PolicyNone,
		rotateCheck: func() bool {
			return false
		},
		rotate: func() bool {
			return true
		},
		newTimer: func() *LogTimer {
			return nil
		},
	}
	// Uses the prefix from LogFile
	lf.filenameGen = lf.getStaticFilename

	lf.Lock()
	defer lf.Unlock()
	err = lf.openFile(lf.filenameGen())
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\", \"timer\":\"%s\"}",
		"start", lf.policy.String(), lf.currentFile, "0")
	log.Printf(msg)
	return
}

// Creates a log file with a size constraint (limit).
// When a file will exceed the size limit, a new volume is created.
//
// The minimum file size is 1MB, and size limits are rounded the next larger multiple.
// The max volume size is set to 500MB.
//
// Given the same prefix, the file is always the same name of "prefix".volNo."log", where volNo
// starts with "0000", and increments. The current volNo will be reopened if it exists, or the next
// one in sequence if the size limit is reached.

//
// If an error occurs, returns nil, and an error.
func SizeLimitedFile(name string, size int64) (lf *LogFile, err error) {
	lf = &LogFile{prefix: name, policy: PolicyFileSize}
	lf.filenameGen = lf.getStaticFilename
	lf.rotateCheck = lf.sizeRotateCheck
	lf.rotate = lf.timedRotate
	lf.newTimer = func() *LogTimer {
		return nil
	}

	size = max(size, LogMaxFileSize)
	if rem := math.Mod(float64(size), float64(LogMinFileSize)); rem > 0.0 {
		size = (size / LogMinFileSize) * LogMinFileSize + LogMinFileSize
	} else {
		size = LogMinFileSize
	}

	lf.fileSizeLimit = size

	lf.Lock()
	defer lf.Unlock()
	err = lf.openFile(lf.filenameGen())
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\", \"size_limit\":\"%d\", \"timer\":\"%s\"}",
		"start", lf.policy.String(), lf.currentFile, lf.fileSizeLimit, "0")
	log.Printf(msg)
	return
}

// Craate a log file using the rotation policy PolicyDaily. There is no size limit for the file.
//
// Creates a file name of prefix "." date ".log", e.g. "appname.2017-01-01.log
// The timer is initialized to rotate at midnight (00:00:00), and reset at each rotation.
// At each file rotation, the file name is updated with the current date.
//
// If an error occurs, then it returns nil, and an error.
func DailyFile(name string) (lf *LogFile, err error) {
	lf = &LogFile{prefix: name, policy: PolicyDaily, cycle: 24 * time.Hour}
	lf.filenameGen = lf.getDailyFilename
	lf.rotateCheck = lf.timedRotateCheck
	lf.rotate = lf.timedRotate

	lf.Lock()
	defer lf.Unlock()
	err = lf.openFile(lf.filenameGen())
	if err != nil {
		lf.Close()
		lf = nil
		return nil, err
	}

	lf.newTimer = func() *LogTimer {
		return NewDailyTimer(time.Now().Location(), func() {
			_ = lf.LogRotate()
		})
	}
	lf.ltimer = lf.newTimer()

	msg := fmt.Sprintf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\", \"timer\":\"%s\"}",
		"start", lf.policy.String(), lf.currentFile, lf.ltimer.d.String())
	log.Printf(msg)
	return lf, nil
}

// Craate a log file using the rotation policy PolicyTimeLimit. There is no size limit for the file.
//
// Creates a file name of "name.YYYY-MM-DDThh_mm_ss.log".
// Name represents a full path and filename prefix.
// The timer is initialized to the current date/time, and reset at each rotation, specified by rt.
// At each file rotation, the file name is updated with the current date and time.
//
// If an error occurs, then it returns nil, and an error.
//
func TimedFile(name string, rt time.Duration) (lf *LogFile, err error) {
	lf = &LogFile{prefix: name, policy: PolicyTimeLimit, cycle: rt}
	lf.filenameGen = lf.getTimedFilename // filename generator
	lf.rotateCheck = lf.timedRotateCheck // Rotation check, true if time
	lf.rotate = lf.timedRotate           // file rotation method

	lf.Lock()
	defer lf.Unlock()
	err = lf.openFile(lf.filenameGen())
	if err != nil {
		lf.Close()
		lf = nil
		return nil, err
	}

	lf.ltimer = NewLocalTimer(lf.cycle, func() {
			_ = lf.LogRotate()
		})

	msg := fmt.Sprintf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\", \"timer\":\"%s\"}",
		"start", lf.policy.String(), lf.currentFile, lf.ltimer.d.String())
	log.Printf(msg)
	return
}

// Return the policy in effect
func (lf *LogFile) LogPolicy() PolicyType {
	return lf.policy
}

// Write a message to the log.  This implements the io.Writer interface
// This is goroutine safe using a mutex lock
func (lf *LogFile) Write(p []byte) (n int, err error) {

	defer func() {
		lf.Unlock()
		if x := recover(); x != nil {
			m := fmt.Sprintf("%s: Error writing to file \"%s\". %s",
				GetCaller(), lf.currentFile, x)
			log.Printf(m)
			n, err = 0, errors.New(m)
			return
		}
		if lf.policy == PolicyFileSize && lf.LogRotateCheck() {
			lf.LogRotate()
		}
	}()
	lf.Lock()

	// strip newlines and add one to the end. Mitigate malformed log events.
	p = append(bytes.Replace(p, []byte("\n"), []byte("; "), -1), '\n')

	n, err = lf.writeEntry(p)
	return
}

// Convenience function.
func (lf *LogFile) writeEntry(p []byte) (n int, err error) {
	n, err = lf.f.Write(p)
	if err != nil {
		log.Printf("%s: %s", GetCaller(), err)
		return 0, err
	}
	return
}

// Close a log file. This implements the io.Closer interface
// If there is a timer associated with the LogFile, Close stops the timer.
// Writes to the log after it is closed may result in an error.
// This is goroutine safe.
func (lf *LogFile) Close() (err error) {
	defer func() {
		lf.Unlock()
		if x := recover(); x != nil {
			log.Printf("Error closing file \"%s\". %s", lf.currentFile, x)
			return
		}
	}()
	lf.Lock()
	if lf.ltimer != nil {
		lf.ltimer.Stop()
	}
	err = lf.f.Close()
	return
}

// Returns the current log file name that is being written calling the FileWriter LogFilename interface.
//
func (lf *LogFile) LogFilename() string {
	return lf.currentFile
}

// Indicates the log file is ready to rotate calling the FileWriter LogRotateCheck interface.
//
// Returns true if it should be rotated, else false. This may be ignored if it is not
// a rotating policy type.
func (lf *LogFile) LogRotateCheck() bool {
	defer func() {
		if x := recover(); x != nil {
			return
		}
	}()
	return lf.rotateCheck()
}

// Rotates the log file calling the FileWriter LogRotate interface.
// Returns true if rotated, false otherwise.
func (lf *LogFile) LogRotate() bool {
	lf.Lock()
	defer lf.Unlock()

	rotated := lf.rotate()
	return rotated
}

// Check for scheduled log file rotation, i.e. PolicyDaily
// Returns true of the rotate time is after the current time.
//
func (lf *LogFile) timedRotateCheck() bool {
	if lf.ltimer == nil {
		return false
	}
	return time.Now().Round(time.Minute).After(lf.ltimer.TriggerTime())
}

// Rotates the log file.
// This generates the new filename, and check if different than the current.
// If it is a new file, then close the old file, and open the new one.
// Returns true if the file was changed, i.e. rotated.
// Assumes the caller synchronizes access.
func (lf *LogFile) timedRotate() (b bool) {
	var dur time.Duration
	log.Printf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\"}",
		"rotate_start", lf.policy.String(), lf.currentFile)

	lf.closeFile()
	lf.openFile(lf.filenameGen())
	b = true

	// If there is a timer, set a new timer.
	if lf.ltimer != nil {
		lf.ltimer.Reset()
		dur = lf.ltimer.Duration()
	}

	msg := fmt.Sprintf("{\"action\":\"%s\", \"policy\":\"%s\", \"file\":\"%s\", \"timer\":\"%s\"}",
		"rotate_end", lf.policy.String(), lf.currentFile, dur)
	log.Printf(msg)
	// Return true, indicating a file change
	return
}

func (lf *LogFile) sizeRotateCheck() bool {
	var ready bool = false
	// Safety check
	if lf.policy != PolicyFileSize {
		return ready
	}

	fi, err := os.Stat(lf.currentFile)
	if err != nil {
		log.Panicf("Error getting log file size for \"%s\". %s.\n",
			lf.currentFile, err)
		// Assume failure, and indicate ready to rotate.
		return true
	}
	ready = ((fi.Size()) + logHighWaterMark > lf.fileSizeLimit)

	return ready
}

// Log File Operations - Open/close.
// Thew NewLogFile() and NewDailyLogFile routines call openFile
// The Close() routine implements the io.Closer interface.

// Open a log file. Called by New
// If successful, returns a nil, else an error.
// The caller must synchronize access.
func (lf *LogFile) openFile(filename string) (err error) {
	lf.f, err = os.OpenFile(filename, logDefaultOpenFlags, logDefaultFileMode)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s: (\"%s\") %s.\n",
			GetCaller(), filename, err))
		return
	}
	lf.currentFile = filename
	return
}

// Close a log file.
// The caller must synchronize access.
func (lf *LogFile) closeFile() (err error) {
	defer func() {
		lf.currentFile = ""
		lf.f = nil
	}()

	if lf.f == nil {
		log.Panicf("Invalid writer. {%s}", lf.currentFile)
	}

	if err = lf.f.Close(); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s: (\"%s\") %s.\n",
			GetCaller(), lf.currentFile, err))
		return err
	}
	return nil
}

// Log file name utilities
//
//

// Generate the filename for the log file
// The "parts" arguments combines with prefix  using the format of
// 	prefix "." part "." part "." ... ".log"
//
// The parts passed to this vary based on the type log log/policy.
func genFilename(prefix string, parts ...string) string {
	fmtStr := prefix + "."
	for _, s := range parts {
		fmtStr += s + "."
	}
	fmtStr += logFilenameExtension
	return fmtStr
}

// Create a static log file name, i.e. PolicyNone, PolicyFileSize
// The filename is prefix "." volume_number ".log".
// Prefix is the path + base filename.
// Volume is a sequence number from 1 to 9999.
func (lf *LogFile) getStaticFilename() string {
	if lf.volNo == 0 {
		lf.volNo = 1
	} else {
		lf.volNo = calcNextVolumeNo(lf.prefix)
	}

	v := int64(lf.volNo)
	s := genFilename(lf.prefix, fmt.Sprintf(logFilenameVolumeFormat, v))
	return s
}

// Calculate the volune number for the next log volume.
// Determines the next number in sequence based by finding the file with the oldeset ModTime,
// extracts the volume number, and then returns the next one in sequence.
// Returns in the range of 1 through 9999. Zero (0) is a reserved volume number.
func calcNextVolumeNo(prefix string) (volNo int16) {
	// Get a list of files
	matches, err := filepath.Glob(genFilename(prefix, "*"))
	if err != nil || matches == nil {
		return 1
	}

	// Find the newest file
	var fi os.FileInfo
	var oldestFile string = matches[0]
	oldestFi, _ := os.Stat(matches[0])
	for _, f := range matches {
		if fi, _ = os.Stat(f); fi.ModTime().After(oldestFi.ModTime()) {
			oldestFi, oldestFile = fi, f
			continue
		}
	}

	// Get the volume number from the filename, and then increment.
	list := regexp.MustCompile("\\.([0-9]+)\\.log").FindAllStringSubmatch(oldestFile, -1)
	n, _ := strconv.ParseInt(list[0][1], 10, 16)
	volNo = int16(math.Mod(float64(n), float64(logMaxVolNumber))) + 1
	if volNo == 0 {
		volNo = 1
	}
	return
}

// Craete a daily log file name, i.e. PolicyDaily
// The filereturned is: prefix "." date ".log"
// the date part takes the form of YYYY-MM-DD.
//
func (lf *LogFile) getDailyFilename() string {
	// Get just the date portion.
	s := time.Now().Format(time.RFC3339)[:len(logFilenameDailyFormat)]
	return genFilename(lf.prefix, s)
}

// Craete a daily log file name, i.e. PolicyTimeLimit.
// The filename includes a date and timestamp. The policy expects the file to be rotated based
// on a set time schedule.
// The file returned is: prefix "." date "T" time ".log". THe ":" in the time is replaced with an
// alternate character. the date part takes the form of YYYY-MM-DDThh_mm_ss.
//
func (lf *LogFile) getTimedFilename() string {
	// Get just the date portion.
	s := time.Now().Format(time.RFC3339)[:len(logFilenameTimeFormat)]
	s = strings.Replace(s, ":", "_", -1)
	return genFilename(lf.prefix, s)
}

func max(x, y int64) (z int64) {
	z = x
	if y > x {
		z = y
	}
	return
}
