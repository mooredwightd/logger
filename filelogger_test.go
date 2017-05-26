package logger

import (
	"flag"
	"fmt"
	"github.com/mooredwightd/gotestutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var longDaily bool

func init() {
	flag.BoolVar(&longDaily, "longdaily", false, "execute the long daily test run")
}

// Success scenarios
// NewStatic=1 : Open a static log, and close
// NewStatic=2 : Open a static log, close, and then reopen (same file name)
func TestFile(t *testing.T) {
	var testName = []string{"test_static01", "test_static02"}
	var testNameBad = []string{"/", "\\//"}

	// Test normal static log creation
	t.Run("NewStatic=1", func(t *testing.T) {
		l, err := File(testName[0])
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s; \"%s\"\n", err, testName))
		gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName[0]))

		p := l.LogPolicy()
		gotestutil.AssertFalse(t, p.IsSizeLimited(), "Expected static file policy, got "+p.String())
		gotestutil.AssertTrue(t, p.isNone(), "Expected Static file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsDaily(), "Expected Static file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsTimed(), "Expected Static file policy, got "+p.String())

		name := l.LogFilename()
		defer os.Remove(name)
		gotestutil.AssertNotEmptyString(t, name, fmt.Sprintf("Value: \"%s\"\n", testName))
		l.Close()

		_, ok := os.Stat(name)
		gotestutil.AssertNil(t, ok, fmt.Sprintf("%s; File: \"%s\"\n", err, testName))
	})

	// Test open/close, reopen
	t.Run("NewStatic=2", func(t *testing.T) {
		l, err := File(testName[1])
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s; File:\"%s\"\n", err, testName))
		gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName[1]))

		p := l.LogPolicy()
		gotestutil.AssertFalse(t, p.IsSizeLimited(), "Expected static file policy, got "+p.String())
		gotestutil.AssertTrue(t, p.isNone(), "Expected Static file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsDaily(), "Expected Static file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsTimed(), "Expected Static file policy, got "+p.String())

		name := l.LogFilename()
		defer os.Remove(name)
		gotestutil.AssertNotEmptyString(t, name, fmt.Sprintf("Value: \"%s\"\n", name))
		l.Close()

		l, err = File(testName[1])
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s; File:\"%s\"\n", err, testName))
		l.Close()
	})

	t.Run("NewStaticBadName=1", func(t *testing.T) {
		for _, s := range testNameBad {
			l, err := File(s)
			if err == nil {
				name := l.LogFilename()
				l.Close()
				t.Fatalf("Fail, created/opened \"%s\"\n", name)
				os.Remove(name)
			}
		}
	})
}

func TestLimitedFile(t *testing.T) {
	testName := "TestLimitedFile"
	var names = make(map[int]string, 5)

	// Test normal static log creation
	runName := testName + "=1"
	t.Run(runName, func(t *testing.T) {
		// Minimum size is 1MB, so it will be rounded to nearest MB
		l, err := SizeLimitedFile(testName, 3*Kbyte)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s; \"%s\"\n", err, testName))
		gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName))

		p := l.LogPolicy()
		gotestutil.AssertTrue(t, p.IsSizeLimited(), "Expected size limited file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.isNone(), "Expected size limited file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsDaily(), "Expected size limited file policy, got "+p.String())
		gotestutil.AssertFalse(t, p.IsTimed(), "Expected size limited file policy, got "+p.String())

		defer func() {
			l.Close()
			for _, v := range names {
				os.Remove(v)
			}
		}()
		names[0] = l.LogFilename()
		gotestutil.AssertNotEmptyString(t, names[0], fmt.Sprintf("Value: \"%s\"\n", testName))
		_, ok := os.Stat(names[0])
		gotestutil.AssertNil(t, ok, fmt.Sprintf("%s; File: \"%s\"\n", err, testName))

		for i := 0; i < 10; i++ {
			msg := strings.Repeat(strconv.Itoa(i), int(256*Kbyte))
			l.Write([]byte(msg))
		}
		names[1] = l.LogFilename()
		gotestutil.AssertStringsNotEqual(t, names[0], names[1], "Expected two different files. "+
			names[0]+" "+names[1])
	})
}

func TestLogFile_Write(t *testing.T) {
	testName := "TestStaticWrite01"

	l, err := File(testName)
	gotestutil.AssertNil(t, err, fmt.Sprintf("%s; File:\"%s\"\n", err, testName))
	gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName))

	name := l.LogFilename()
	defer func() {
		l.Close()
		os.Remove(name)
	}()

	testMsg := "Test log message."
	_, wErr := l.Write([]byte(testMsg))
	if wErr != nil {
		t.Fatalf("%s: Error on write.\n", GetCaller())
	}
}

func TestTimedFile(t *testing.T) {
	var timeSpan = 1 * time.Minute
	var numRuns = 3
	testName := "TestTimedLog"

	for i := 0; i < numRuns; i++ {
		t.Run(testName, func(t *testing.T) {
			log.Printf("Start run %d...\n", i)

			// Set the scheduled rotation time
			l, err := TimedFile(fmt.Sprintf("%s%2.2d", testName, i), timeSpan)
			gotestutil.AssertNil(t, err, fmt.Sprintf("File \"%s\"\n", testName))
			gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName))

			p := l.LogPolicy()
			gotestutil.AssertFalse(t, p.IsSizeLimited(), "Expected timed file policy, got "+p.String())
			gotestutil.AssertFalse(t, p.isNone(), "Expected timed file policy, got "+p.String())
			gotestutil.AssertFalse(t, p.IsDaily(), "Expected timed file policy, got "+p.String())
			gotestutil.AssertTrue(t, p.IsTimed(), "Expected timed file policy, got "+p.String())

			name1 := l.LogFilename()
			gotestutil.AssertNotEmptyString(t, name1, "Filename: "+name1)

			l.Write([]byte(fmt.Sprintf("Message, Line 1 - Run %d", i)))
			time.Sleep(timeSpan + (7 * time.Second)) // Sleep past the schedule time
			l.Write([]byte(fmt.Sprintf("Message, Line 2 - Run %d", i)))

			name2 := l.LogFilename() // Filename should be different
			gotestutil.AssertNotEmptyString(t, name2, "Filename: "+name2)

			gotestutil.AssertStringsNotEqual(t, name1, name2,
				fmt.Sprintf("File 1: \"%s\" File 2 \"%s\".\n", name1, name2))

			_, ok1 := os.Stat(name1)
			_, ok2 := os.Stat(name2)
			gotestutil.AssertNil(t, ok1, fmt.Sprintf("%s; File: \"%s\".", ok1, name1))
			gotestutil.AssertNil(t, ok2, fmt.Sprintf("%s; File: \"%s\".", ok2, name2))
			log.Printf("Finishing run %d...\n", i)
			l.Close()
			os.Remove(name1)
			os.Remove(name2)
		})
	}

}

func TestDailyFile(t *testing.T) {
	testName := "TestDailyLog01"

	l, err := DailyFile(testName)
	gotestutil.AssertNil(t, err, fmt.Sprintf("Error opening \"%s\"\n", testName))
	gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName))

	p := l.LogPolicy()
	gotestutil.AssertFalse(t, p.IsSizeLimited(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertFalse(t, p.isNone(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertTrue(t, p.IsDaily(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertFalse(t, p.IsTimed(), "Expected Daily file policy, got "+p.String())

	name1 := l.LogFilename()
	defer os.Remove(name1)

	l.Write([]byte("Message, Line 1 - "))
	l.Write([]byte("Message, Line 2 - "))
	l.Close()

	_, ok1 := os.Stat(name1)
	gotestutil.AssertNil(t, ok1, fmt.Sprintf("%s; File: \"%s\".", ok1, name1))
}

func TestDailyFile2(t *testing.T) {
	testName := "TestDailyLog02"

	if !longDaily {
		t.Skip("Skipping TestDailyLogFile (TestDailyLog02)")
	}

	l, err := DailyFile(testName)
	gotestutil.AssertNil(t, err, fmt.Sprintf("Error opening \"%s\"\n", testName))
	gotestutil.AssertNotNil(t, l, fmt.Sprintf("*LogFile is nil: \"%s\"\n", testName))

	p := l.LogPolicy()
	gotestutil.AssertFalse(t, p.IsSizeLimited(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertFalse(t, p.isNone(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertTrue(t, p.IsDaily(), "Expected Daily file policy, got "+p.String())
	gotestutil.AssertFalse(t, p.IsTimed(), "Expected Daily file policy, got "+p.String())

	name1 := l.LogFilename()

	l.Write([]byte("Message, Line 1 - "))
	n := time.Now()
	mn := time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, n.Location())
	d := mn.Sub(n)
	log.Printf("Sleep until midnight. %s", d.String())
	time.Sleep(d)
	l.Write([]byte("Message, Line 2 - "))

	name2 := l.LogFilename()
	defer os.Remove(name1)
	defer os.Remove(name2)

	l.Close()

	_, ok1 := os.Stat(name1)
	_, ok2 := os.Stat(name2)
	gotestutil.AssertNil(t, ok1, fmt.Sprintf("%s; File: \"%s\".", ok1, name1))
	gotestutil.AssertNil(t, ok2, fmt.Sprintf("%s; File: \"%s\".", ok2, name2))

}
