package logger

import (
	"strings"
	"fmt"
)

const (
	DefaultFieldSeparator string = "|"	// For log formatter
)

type PlainTextFormatter struct {
	name      string
	separator string
}

// Create a new Plain Text event message formatter.
// Returns an EventFormatter interface.
func PlainText() EventFormatter {
	return PlainTextFormatter{
		name:"plain_text",
		separator:DefaultFieldSeparator}
}

// Set the field delimeter for log messages.
func (ptf PlainTextFormatter) SetDelimeter(d string) {
	ptf.separator = d
}

// Implements EventFormatter interface.
func (ptf PlainTextFormatter) Format(em EventMsg) (msg string, err error) {
	var defaultFmt = "%s|%s|%s|%s|%d|%s|%s|"
	var altSepFmt = "%s%s%s%s%s%s%s%s%d%s%s%s%s%s"
	ptf.separator = DefaultFieldSeparator

	tm := strings.Replace(fmt.Sprintf("%s", em.Timestamp.String()), " ", "", -1)
	if ptf.separator == DefaultFieldSeparator {
		msg = fmt.Sprintf(defaultFmt,
			tm, em.Sev, em.Hostname, em.Appname, em.Pid, em.MsgId, em.Msg)
	} else {
		msg = fmt.Sprintf(altSepFmt,
			tm, ptf.separator,
			em.Sev, ptf.separator,
			em.Hostname, ptf.separator,
			em.Appname,ptf.separator,
			em.Pid, ptf.separator,
			em.MsgId, ptf.separator,
			em.Msg, ptf.separator)
	}
	msg += "["
	for n, v := range em.Params {
		msg += fmt.Sprintf("%s=%s,", n, v)
	}
	if msg[len(msg) - 1] == ',' {
		msg = msg[:len(msg) - 1]
	}
	msg += "]"
	return
}
