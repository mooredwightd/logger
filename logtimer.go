// LogTimer is used for time or interval based log rotations using Go time functions.
// Default implementation is
//     DailyTimer, via NewDailyTImer, that establishes a timer that fires at 00:00:00.
//     Timer, via NewTimer, which creates a generic timer that fires after the specified duration.
package logger

import (
	"log"
	"time"
)

// This is the base structure for a timer. It augments the Go time.Timer.
type LogTimer struct {
	base  time.Time     // Base time of when the timer is started.
	next  time.Time     // Should be base + duration
	d     time.Duration // The duration registered when the timer was started.
	cb    func()        // provided by client/caller
	timer *time.Timer   // Pointer to the Go Timer
}

// Create a new timer that executes the function parameter at the given time.
// This timer starts the basetime at 12am (midnight) based on the location specified.
// The duration is always calculated as the difference  between now and 12am.
func NewDailyTimer(loc *time.Location, f func()) (lt *LogTimer) {
	lt = &LogTimer{d: 24 * time.Hour, cb: f}
	t := time.Now()
	if loc == nil {
		loc = t.Location()
	}
	lt.base = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Round(time.Minute)
	d := lt.calcDurationFromNow(lt.base.Add(lt.d))

	lt.next = lt.base.Add(lt.d)
	lt.timer = time.AfterFunc(d, lt.doTimerFunc)
	return lt
}

// Creates a new timer for the "local" location (current time zone).
// The duration specifies when the timer executes the function parameter (f).
// This is the same as calling NewTimer with time.Now().Location.
func NewLocalTimer(dur time.Duration, f func()) (lt *LogTimer) {
	l, _ := time.LoadLocation("Local")
	return NewTimer(dur, l, f)
}

// Create a new timer to execute a function at the specified time.
// The dur is the duration to wait before calling the function.
// The location is the time zone for the timer (clock).
// The function, f, is  called when the timer expires.
func NewTimer(dur time.Duration, l *time.Location, f func()) (lt *LogTimer) {
	lt = &LogTimer{d: dur, cb: f}

	n := time.Now()
	if l == nil {
		l = n.Location()
	}
	lt.base = time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), 0, 0, l)
	lt.next = lt.base.Add(lt.d)
	lt.timer = time.AfterFunc(lt.d, lt.doTimerFunc)
	return
}

// Stop the timer.
// If the timer has stopped or expired, it drains the channel.
func (lt *LogTimer) Stop() {
	if lt.timer == nil {
		return
	}
	if !lt.timer.Stop() && lt.timer.C != nil {
		<-lt.timer.C
	}
}

// Reset the timer by stopping, and then reset the duration, starting an active timer.
func (lt *LogTimer) Reset() {
	lt.Stop()
	lt.timer.Reset(lt.d)
}

// Returns the duration of the timer
func (lt *LogTimer) Duration() (d time.Duration) {
	return lt.d
}

// Returns the trigger time, i.e. the time the callback is expected to be called.
func (lt *LogTimer) TriggerTime() time.Time {
	return lt.next
}

// Returns the current time location.
func (lt *LogTimer) Location() *time.Location {
	return lt.base.Location()
}

// Internal timer callback. This calls the registered function when the timer was created.
// This allows for other actions to be wrapped around it, i.e. before/after the callback.
func (lt *LogTimer) doTimerFunc() {
	defer func() {
		if x := recover(); x != nil {
			log.Printf("LogTimer: panic during in doTimerFunc(). %s.\n", x)
		}
	}()
	lt.cb()
	lt.base = lt.next
}

// Calculate a duration beteen now and a future time.
// Rounds to the neareset minute.
func (lt *LogTimer) calcDurationFromNow(t time.Time) (d time.Duration) {
	d = t.Sub(time.Now().Round(time.Minute))
	//log.Printf("calcDurationFromNow: now:%s, future:%s, duration:%s", time.Now().String(), t.String(), d.String())
	return
}
