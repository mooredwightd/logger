package logger

import "strings"

// Public constants
type Severity int16

// Severity constants
const (
	InvalidSeverity = -1
	Emergency       = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Info
	Debug

	SeverityMinLevel = Emergency
	SeverityMaxLevel = Debug
)

var (
	// Text representation of log levels
	severityToString = [...]string{
		"Invalid", "EMERG", "ALERT", "CRIT", "ERROR", "WARN", "NOTIC", "INFO", "DEBUG",
	}
	stringToSeverity = map[string]int{
		severityToString[Emergency]: Emergency,
		severityToString[Alert]:     Alert,
		severityToString[Critical]:  Critical,
		severityToString[Error]:     Error,
		severityToString[Warning]:   Warning,
		severityToString[Notice]:    Notice,
		severityToString[Info]:      Info,
		severityToString[Debug]:     Debug,
	}
)


// Returns the string representation of a Severity value.
// Returns true if valid, else false.
func (s Severity) String() string {
	return severityToString[s]
}

// Validates if a string represents a severity level.
func IsValidSeverity(s string) bool {
	_, valid := stringToSeverity[strings.ToUpper(s)]
	return valid
}

// Translates a text string to a Severity.
// If the text string is not valid, returns InvalidSeverity
func StringToSeverity(s string) Severity {
	v, valid := stringToSeverity[strings.ToUpper(s)]
	if !valid {
		return InvalidSeverity
	}
	return Severity(v)
}

