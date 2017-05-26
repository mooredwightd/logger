package logger

import (
	"fmt"
	"log"
	"testing"
	"time"
	"github.com/mooredwightd/gotestutil"
)

const (
	MaxIterations = 3
	CallbackDuration = 10 * time.Second
	SleepDuration = 5 * time.Second
)

type TestStruct struct {
	name     string
	tStamp   time.Time
	c        chan string
	cb       func()
	received bool
	msgCnt   int
}

type OtherStruct struct {
	Txt string
}

var osTest = OtherStruct{"OtherStruct"}

func NewTS(n string, d time.Duration, f func()) *TestStruct {
	nt := time.Now()
	t := &TestStruct{name: n, tStamp: time.Date(nt.Year(), nt.Month(), nt.Day(),
		nt.Hour(), nt.Minute(), nt.Second(), 1, time.Now().Location())}
	t.c = make(chan string)
	t.cb = f
	return t
}

func (t *TestStruct) testCallBack() {
	msg := fmt.Sprintf("Call back called. %s, %s, count %d", t.name, t.tStamp, t.msgCnt)
	log.Println(msg)
	t.cb()
	t.tStamp = time.Now()
	t.c <- msg
}

func (os *OtherStruct) innerCallback() {
	log.Printf("Inner callback called: %s", os.Txt)
}

func TestLogTimer_Duration(t *testing.T) {
	tmr := NewTimer(CallbackDuration, time.Now().Location(), func() {
		log.Println("TestLogTImer_Duration: callback executed.")
	})
	d := tmr.Duration()
	gotestutil.AssertEqual(t, CallbackDuration, d, "Durations did not match.")
	tmr.Stop()
}

func TestLogTimer_TriggerTime(t *testing.T) {
	tmr := NewTimer(time.Minute * 5, time.Now().Location(), func() {
		log.Println("TestLogTimer_TriggerTime: callback executed.")
	})
	defer tmr.Stop()
	n := tmr.TriggerTime()
	if time.Now().After(tmr.TriggerTime()) {
		t.Fatalf("Trigger time is earlier than current time. Trigger at %s", n.String())
	}
}

func TestNewTimer(t *testing.T) {
	var msg string
	name1 := "NewTimer01"
	ts1 := NewTS(name1, CallbackDuration, osTest.innerCallback)
	tmr := NewTimer(CallbackDuration, time.Now().Location(), ts1.testCallBack)

	for done := false; !done; {
		select {
		case msg = <-ts1.c:
			log.Printf("%s Msg received. \"%s\"", name1, msg)
			ts1.received = true
			ts1.msgCnt++
			if ts1.msgCnt < MaxIterations {
				log.Printf("%s Reset timer. %d", name1, ts1.msgCnt)
				tmr.Reset()
			} else {
				done = true
			}
		default:
			var delay time.Duration = SleepDuration
			log.Printf("%s Sleep %4.4f seconds...", name1, delay.Seconds())
			time.Sleep(delay)
		}
	}
	tmr.Stop()
	gotestutil.AssertTrue(t, ts1.received, "TestNewTimer: No message received.")
	gotestutil.AssertEqual(t, ts1.msgCnt, MaxIterations, "TestNewTimer: did not receive all messages")
}

func TestNewLocalTimer(t *testing.T) {
	var msg string
	name1 := "NewLocalTimer01"
	ts1 := NewTS(name1, CallbackDuration, osTest.innerCallback)
	tmr := NewLocalTimer(CallbackDuration, ts1.testCallBack)
	_ = tmr

	var done bool = false
	// Stop an infinite loop....just in case. Must execute within the duration
	loopStop := time.AfterFunc(3 * time.Minute, func() {
		log.Println("Loop stop. Too much time.")
		done = true
	})

	for !done {
		select {
		case msg = <-ts1.c:
			ts1.received = true
			log.Printf("%s message received. Count %d. %s", name1, ts1.msgCnt, msg)
			ts1.msgCnt++
			if ts1.msgCnt < MaxIterations {
				log.Printf("%s Reset timer. %d", name1, ts1.msgCnt)
				tmr.Reset()
			} else {
				done = true
			}
		default:
			var delay time.Duration = CallbackDuration
			log.Printf("%s Sleep %4.4f seconds...", name1, delay.Seconds())
			time.Sleep(delay)
		}
	}
	loopStop.Stop()
	tmr.Stop()
	gotestutil.AssertTrue(t, ts1.received, "TestNewLocalTimer: No message received.")
	gotestutil.AssertEqual(t, ts1.msgCnt, MaxIterations,
		"TestNewLocalTimer: msgCnt less than expected. Expected: %d, Actual: %d",
		MaxIterations, ts1.msgCnt)
}

func TestLogTimer_Stop(t *testing.T) {
	var msg string
	name1 := "StopTimer"
	ts1 := NewTS(name1, CallbackDuration, osTest.innerCallback)
	tmr := NewLocalTimer(CallbackDuration, ts1.testCallBack)
	var done bool = false
	_ = msg

	time.Sleep(SleepDuration)
	tmr.Stop()
	for !done {
		select {
		case msg = <-ts1.c:
			t.Logf("%s message received. Count %d. %s", name1, ts1.msgCnt, msg)
			ts1.received = true
			ts1.msgCnt++
			if ts1.msgCnt <= MaxIterations {
				log.Printf("%s Reset timer. %d", name1, ts1.msgCnt)
				tmr.Reset()
			} else {
				done = true
			}
		default:
			done = true
		}
	}
	gotestutil.AssertFalse(t, ts1.received, "TestLogTimer_Stop: No message received.")
}

// Benchmark tests
func BenchmarkNewTimer(b *testing.B) {
	name1 := "NewTimer01"
	ts1 := NewTS(name1, CallbackDuration, osTest.innerCallback)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmr := NewTimer(CallbackDuration, time.Now().Location(), ts1.testCallBack)
		tmr.Stop()
	}
	b.StopTimer()
}

func BenchmarkNewDailyTimer(b *testing.B) {
	name1 := "NewTimer01"
	ts1 := NewTS(name1, CallbackDuration, osTest.innerCallback)
	for i := 0; i < b.N; i++ {
		tmr := NewDailyTimer(time.Now().Location(), ts1.testCallBack)
		tmr.Stop()
	}
}
