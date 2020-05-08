package logger

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/mooredwightd/gotestutil"
)

var verbose bool

func init() {
	flag.BoolVar(&verbose, "verbose", false, "execute more verbose logging")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func checkForJsonFields(t *testing.T, fmap map[int]string) (found bool) {
	found = true
	flds1 := []string{"timestamp", "severity", "hostname", "appname",
		"pid", "msg_id", "message", "params"}
	for _, v := range flds1 {
		isThere := gotestutil.AssertTextInFiles(t, fmap, v)
		found = (isThere && found)
	}
	return
}

func checkStringForJsonFields(s string) (found bool) {
	found = true
	var flds = [...]string{"timestamp", "severity", "hostname",
		"pid", "msg_id", "message", "params"}
	for _, v := range flds {
		ok := strings.Contains(s, v)
		found = (found && ok)
		if !found {
			log.Printf("%s: missing field %s", GetCaller(), v)
		}
	}
	return
}

func TestSeverity_String(t *testing.T) {
	testName := "TestSeverity_String"

	runName := testName + "=1"

	t.Run(runName, func(t *testing.T) {
		v := Severity(Emergency).String()
		gotestutil.AssertEqual(t, severityToString[Emergency], v, GetCaller()+" Expected EMERG")
		v = Severity(Alert).String()
		gotestutil.AssertEqual(t, severityToString[Alert], v, GetCaller()+" Expected ALERT")
		v = Severity(Critical).String()
		gotestutil.AssertEqual(t, severityToString[Critical], v, GetCaller()+" Expected CRIT")
		v = Severity(Error).String()
		gotestutil.AssertEqual(t, severityToString[Error], v, GetCaller()+" Expected ERROR")
		v = Severity(Warning).String()
		gotestutil.AssertEqual(t, severityToString[Warning], v, GetCaller()+" Expected WARN")
		v = Severity(Notice).String()
		gotestutil.AssertEqual(t, severityToString[Notice], v, GetCaller()+" Expected NOTIC")
		v = Severity(Info).String()
		gotestutil.AssertEqual(t, severityToString[Info], v, GetCaller()+" Expected INFO")
		v = Severity(Debug).String()
		gotestutil.AssertEqual(t, severityToString[Debug], v, GetCaller()+" Expected DEBUG")
	})

	// Should fail for invalid Severity type.
	// Throws a panic
	runName = testName + "=2"
	t.Run(runName, func(t *testing.T) {
		v := Severity(0).String()
		gotestutil.AssertEqual(t, severityToString[0], v, GetCaller()+" Expected panic")
	})

	// Should fail for invalid Severity type.
	// Throws a panic
	runName = testName + "=3"
	t.Run(runName, func(t *testing.T) {
		defer func() {
			if x := recover(); x != nil {
				log.Printf("%s caught panic %s", runName, x)
				return
			}
		}()

		v := Severity(100).String()
		gotestutil.AssertEqual(t, "", v, GetCaller()+" Expected panic")
	})

}

func TestIsValidSeverity(t *testing.T) {
	testName := "TestIsValidSeverity"
	var ok bool

	runName := testName + "=1"
	t.Run(runName, func(t *testing.T) {
		var tStr = [...]string{"EMERG", "ALERT", "CRIT", "ERROR", "WARN", "NOTIC",
			"INFO", "DEBUG"}
		for _, v := range tStr {
			ok = IsValidSeverity(v)
			gotestutil.AssertTrue(t, ok, runName+" Expected true for "+v)
		}
	})

	runName = testName + "=2"
	t.Run(runName, func(t *testing.T) {
		var tStr = [...]string{"", "BLAHH", "INVALID", ""}
		for _, v := range tStr {
			ok = IsValidSeverity(v)
			gotestutil.AssertFalse(t, ok, runName+" Expected true for "+v)
		}
	})

}

func TestLogManger(t *testing.T) {
	testName := "TestNewLogManager01"

	// Happy path
	t.Run(testName+"=1", func(t *testing.T) {
		lf, err := File(testName)
		gotestutil.AssertNil(t, err, "Error creating new log file w/ NewFile()")
		gotestutil.AssertNotNil(t, lf, "Expected valid LogFile")
		fn := lf.LogFilename()

		l := LogManger("TestNew", lf)
		defer func() {
			l.Close()
			os.Remove(fn)
		}()

		tStr := "TestNew01 test msg"
		l.Alert(testName, tStr, map[string]string{})
		cOk := gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, tStr)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+tStr)
		cOk = checkForJsonFields(t, map[int]string{1: fn})
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected JSON tags not found.")
	})

	// Invalid LogFile paramter for New File
	// Alert should panic
	t.Run(testName+"=2", func(t *testing.T) {
		//var except = false
		defer func() {
			if x := recover(); x != nil {
				log.Printf("%s Caught exception", testName+"=2")
				//except = true
				t.Fatalf("%s failed", testName+"=2")
				return
			}
		}()
		lf := &LogFile{}
		l := LogManger("TestNew", lf)
		l.Alert(testName, "TestNew01 test msg", map[string]string{})
	})
}

func TestLog_AddLogger(t *testing.T) {
	testName := "TestLog_AddLogger"
	tStr := testName + " test string"
	var l *Log
	var names = make(map[int]string, 5)

	runName := testName + "=1"
	t.Run(runName, func(t *testing.T) {
		defer func() {
			for _, v := range names {
				os.Remove(v)
			}
		}()
		name := fmt.Sprintf("%s_%02.2d", testName, 1)
		lf, err := File(name)
		gotestutil.AssertNil(t, err, GetCaller()+" Error creating new log file w/ NewFile()")
		gotestutil.AssertNotNil(t, lf, GetCaller()+" Expected valid LogFile")
		names[1] = lf.LogFilename()
		l = LogManger(testName, lf)
		l.SetFilter(Debug)
		defer func() {
			l.Close()
		}()

		for i := 2; i <= 3; i++ {
			name := fmt.Sprintf("%s%02.2d", testName, i)
			lf, err := File(name)
			gotestutil.AssertNil(t, err, GetCaller()+" Error creating new log file w/ NewFile()")
			gotestutil.AssertNotNil(t, lf, GetCaller()+" Expected valid LogFile")
			l.AddLogger(lf)
			names[i] = lf.LogFilename()
		}
		l.Info(testName, tStr, map[string]string{})
		gotestutil.AssertEqual(t, len(l.logModules), 3, GetCaller()+" Expected 3 loggers")

		cOk := gotestutil.AssertTextInFiles(t, names, tStr)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+tStr)
		cOk = checkForJsonFields(t, names)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected JSON tags not found.")
	})

	// Insert an invalid LogFile into the Log Manager
	// Should log an error, but not fail.
	names = make(map[int]string, 5)
	runName = testName + "=2"
	t.Run(runName, func(t *testing.T) {
		defer func() {
			for _, v := range names {
				os.Remove(v)
			}
		}()
		lf := &LogFile{}
		l = LogManger(testName, lf)
		l.SetFilter(Debug)
		defer func() {
			l.Close()
		}()
		names[1] = lf.LogFilename()

		for i := 2; i <= 3; i++ {
			name := fmt.Sprintf("%s%02.2d", testName, i)
			lf, err := File(name)
			gotestutil.AssertNil(t, err, GetCaller()+" Error creating new log file w/ NewFile()")
			gotestutil.AssertNotNil(t, lf, GetCaller()+" Expected valid LogFile")
			l.AddLogger(lf)
			names[i] = lf.LogFilename()
		}
		tStr := testName + " test msg"
		l.Alert(testName, tStr, map[string]string{})
		gotestutil.AssertEqual(t, len(l.logModules), 3, GetCaller()+" Expected 3 loggers")

		cOk := gotestutil.AssertTextInFiles(t, names, tStr)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+tStr)
		cOk = checkForJsonFields(t, names)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected JSON tags not found.")
	})
}

func TestLog_Close(t *testing.T) {
	testName := "TestLog_Close"

	t.Run(testName+"=1", func(t *testing.T) {
		name := testName + "01"
		lf, err := File(name)
		gotestutil.AssertNil(t, err, GetCaller()+" Error creating new log file w/ NewFile()")
		gotestutil.AssertNotNil(t, lf, GetCaller()+" Expected valid LogFile")

		fn := lf.LogFilename()
		l := LogManger(testName, lf)
		defer func() {
			os.Remove(fn)
			l.Close()
		}()
		gotestutil.AssertEqual(t, len(l.logModules), 1, GetCaller()+" Expected 1 logger")
		l.Close()
		gotestutil.AssertEqual(t, len(l.logModules), 0, GetCaller()+" Expected 0 logger")
	})

	// Create an invalid file to close
	t.Run(testName+"=2", func(t *testing.T) {
		lf := &LogFile{}
		l := LogManger(testName, lf)
		l.Close()
		gotestutil.AssertEqual(t, len(l.logModules), 0, GetCaller()+" Expected 0 loggers")
	})
}

func TestLog_LogEvent(t *testing.T) {
	testName := "TestLog_LogEvent"

	t.Run(testName+"=1", func(t *testing.T) {
		name := testName + "1"
		lf, err := File(name)
		gotestutil.AssertNil(t, err, GetCaller()+" Error creating new log file w/ NewFile()")
		gotestutil.AssertNotNil(t, lf, GetCaller()+" Expected valid LogFile")

		fn := lf.LogFilename()
		l := LogManger(testName, lf)
		l.SetFilter(Debug)
		defer func() {
			l.Close()
			os.Remove(fn)
		}()
		gotestutil.AssertEqual(t, len(l.logModules), 1, GetCaller()+" Expected 1 logger")

		tStr := testName + " log event test message."
		params := map[string]string{
			"p1": "param1",
			"p2": "param2",
		}
		l.LogEvent(Info, testName, tStr, params)

		cOk := gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, tStr)
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+tStr)

		cOk = checkForJsonFields(t, map[int]string{1: fn})
		gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected JSON tags not found.")
		for k, v := range params {
			cOk = gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, k)
			gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+k)
			cOk = gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, v)
			gotestutil.AssertTrue(t, cOk, GetCaller()+" Expected string not found: "+v)
		}

	})
}

func testSeverities(t *testing.T, testName string, sev Severity) (success bool) {
	lf, _ := File(testName)
	fn := lf.LogFilename()
	l := LogManger(testName, lf)
	l.SetFilter(Debug)
	defer func() {
		l.Close()
		os.Remove(fn)
	}()
	var tStr string

	switch sev {
	case Emergency:
		l.Emergency(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Emergency).String()
	case Alert:
		l.Alert(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Alert).String()
	case Critical:
		l.Critical(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Critical).String()
	case Error:
		l.Error(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Error).String()
	case Warning:
		l.Warning(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Warning).String()
	case Notice:
		l.Notice(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Notice).String()
	case Info:
		l.Info(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Info).String()
	case Debug:
		l.Debug(testName, testName+" test message.", map[string]string{})
		tStr = Severity(Debug).String()
	default:
		tStr = "InvalidSeverityString"
	}
	success = gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, tStr)
	return
}

func TestLog_Emergency(t *testing.T) {
	runName := "TestLog_Emergency"
	stype := Severity(Emergency)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Alert(t *testing.T) {
	runName := "TestLog_Alert"
	stype := Severity(Alert)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Critical(t *testing.T) {
	runName := "TestLog_Critical"
	stype := Severity(Critical)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Error(t *testing.T) {
	runName := "TestLog_Error"
	stype := Severity(Error)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Warning(t *testing.T) {
	runName := "TestLog_Warning"
	stype := Severity(Warning)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Notice(t *testing.T) {
	runName := "TestLog_Notice"
	stype := Severity(Notice)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Info(t *testing.T) {
	runName := "TestLog_Info"
	stype := Severity(Info)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

func TestLog_Debug(t *testing.T) {
	runName := "TestLog_Debug"
	stype := Severity(Debug)
	ok := testSeverities(t, runName, stype)
	gotestutil.AssertTrue(t, ok, GetCaller()+" Expected "+stype.String())
}

// Test that the message is filtered.
func TestLog_Debug2(t *testing.T) {
	testName := "TestLog_Debug2"
	lf, _ := File(testName)
	fn := lf.LogFilename()
	// Default filter is WARN
	l := LogManger(testName, lf)
	defer func() {
		l.Close()
		os.Remove(fn)
	}()
	lStr := testName + " test message for filter"

	l.SetFilter(Alert)
	l.Alert(testName, lStr, map[string]string{})
	l.Debug(testName, lStr, map[string]string{})

	success := gotestutil.AssertTextInFiles(t, map[int]string{1: fn}, Severity(Alert).String())
	gotestutil.AssertTrue(t, success, GetCaller()+" Expected to find unfiltered entry.")

	tStr := Severity(Debug).String()
	success = gotestutil.AssertTextNotInFiles(t, map[int]string{1: fn}, tStr)
}
