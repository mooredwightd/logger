// Formatter provides formatting for log events using the EventFormatter interface.
// A custom formatter is created with the EventFormatter interface, ans set with the
// SetFormatter() method of the LogManager.
//
// The base implementation provides a simple JSONFormatter which formats the event messsage.
// The "plaintext" formatter formats a simple text format with a default "|" (horizontal bar)
// separator, that can be changes with the SetDelimiter() method of the PlainTextFormatter
package logger

import (
	"net"
	"os"
	"strings"
	"time"
)

type EventFormatter interface {
	Format(em EventMsg) (string, error)
}

// Set default values, and validate severity, hostname pid, and trim text.
func validateEventMsg(em *EventMsg) *EventMsg {
	if em.Timestamp.IsZero() || em.Timestamp.Year() != time.Now().Year() {
		em.Timestamp = time.Now().Round(time.Millisecond)
	}
	if !IsValidSeverity(em.Sev) {
		em.Sev = Severity(Warning).String()
		em.Msg += " (Invalid severity in log event specified: " + em.Sev + ")"
	}
	if len(em.Hostname) == 0 {
		if h, hErr := os.Hostname(); hErr != nil {
			if a, aErr := net.InterfaceAddrs(); aErr != nil {
				em.Hostname = "unknown: "
			} else {
				em.Hostname = a[0].String()
			}
		} else {
			em.Hostname = h
		}
	}
	em.Hostname = strings.TrimSpace(em.Hostname)
	em.Appname = strings.TrimSpace(em.Appname)
	if em.Pid == 0 {
		em.Pid = os.Getpid()
	}
	return em
}

// Create a timestamp that is compliant with RFC 5424
// Examples:
//   1985-04-12T23:20:50.52Z => 20 minutes and 50.52 seconds after the 23rd hour of
//     12 April 1985 in UTC.
//   1985-04-12T19:20:50.52-04:00 => Same as above but in US EST (observing daylight savings time).
//   2003-10-11T22:14:15.003Z => UTC with 3 milliseconds into the next second.
///  2003-08-24T05:14:15.000003-07:00 => -7 hours from UTC, with 3 milleseconds ino the next sec.
func timestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
