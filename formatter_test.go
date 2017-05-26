package logger

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
	"github.com/mooredwightd/gotestutil"
)

var emBase EventMsg

func init() {
	emBase.Timestamp = time.Now()
	emBase.Sev = Severity(Info).String()
	hn, _ := os.Hostname()
	emBase.Hostname = strings.TrimSpace(hn)
	emBase.Appname = "formatter_test"
	emBase.MsgId = "MsgId_1"
	emBase.Pid = os.Getpid()
	emBase.Msg = "Test message."
	emBase.Params = map[string]string{
		"p1": "param1",
		"p2": "param2",
		"p3": "param3",
	}
}

// Normal case
func TestJsonFormat(t *testing.T) {
	var em EventMsg
	jf := Json()

	t.Run("A=1", func(t *testing.T) {
		// Normal
		em = emBase
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=2", func(t *testing.T) {
		// empty struct
		em = EventMsg{}
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=3", func(t *testing.T) {
		// empty w/ Timestamp
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=4", func(t *testing.T) {
		// empty w/ Timestamp & severity
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		em.Sev = em.Sev
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		fmt.Printf("%s\n", m)
	})
	t.Run("A=5", func(t *testing.T) {
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		em.Sev = em.Sev
		em.Hostname = emBase.Hostname
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=6", func(t *testing.T) {
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		em.Sev = em.Sev
		em.Hostname = emBase.Hostname
		em.Appname = emBase.Appname
		em.Pid = emBase.Pid
		em.MsgId = emBase.MsgId
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=7", func(t *testing.T) {
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		em.Sev = em.Sev
		em.Hostname = emBase.Hostname
		em.Appname = emBase.Appname
		em.Pid = emBase.Pid
		em.MsgId = emBase.MsgId
		em.Msg = emBase.Msg
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("A=8", func(t *testing.T) {
		em = EventMsg{}
		em.Timestamp = emBase.Timestamp
		em.Sev = em.Sev
		em.Hostname = emBase.Hostname
		em.Appname = emBase.Appname
		em.Pid = emBase.Pid
		em.MsgId = emBase.MsgId
		em.Msg = emBase.Msg
		em.Params = emBase.Params
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
}

func TestJsonFormat2(t *testing.T) {
	var em EventMsg
	jf := Json()

	fmt.Println()
	t.Run("B=1", func(t *testing.T) {
		// Invalid Sev
		em = emBase
		em.Sev = "invalid severity"
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	t.Run("B=2", func(t *testing.T) {
		// Invalid year for timestamp
		em = emBase
		n := emBase.Timestamp
		em.Timestamp = time.Date(1970, n.Month(), n.Day(), n.Hour(), n.Minute(),
			n.Second(), 0, n.Location())
		m, err := jf.Format(em)
		gotestutil.AssertNil(t, err, fmt.Sprintf("%s\n", err))
		gotestutil.AssertGreaterThan(t, len(m), 0, "Message is empty")
		fmt.Printf("%s\n", m)
	})
	fmt.Println()
}

func BenchmarkJsonFormat(b *testing.B) {
	em := emBase
	jf := Json()

	for i := 0; i < b.N; i++ {
		m, err := jf.Format(em)
		_, _ = m, err
	}
}
