// Policy - The log file policy determines the file rotation parameters.
//
package logger

type PolicyType uint64

const (
	invalidPolicy = iota
	// No policy. A single file that does not rotate.
	PolicyNone
	// Daily rotation
	PolicyDaily
	// Rotate based on a given time duration, e.g. every 8 hours.
	PolicyTimeLimit
	// Rotate based on the file size.
	PolicyFileSize
	// For future expansion
	PolicyCustom1
	PolicyCustom2
	PolicyCustom3
)

// Sring representation of the policy
var policyName = []string{
	"Invalid", "PolicyNone", "PolicyDaily", "PolicyTimeLimit", "PolicyFileSize",
}

// Returns the string representation of the policy
func (pt PolicyType) String() string {
	return policyName[pt]
}

// Returns true if the log file has no rotation policy
func (pt PolicyType) isNone() bool {
	return (pt == PolicyNone)
}

// Returns true if the log file has daily rotation policy
func (pt PolicyType) IsDaily() bool {
	return (pt == PolicyDaily)
}

// Returns true if the log file has timed/scheduled rotation policy
func (pt PolicyType) IsTimed() bool {
	return (pt == PolicyTimeLimit)
}

// Returns true if the log file has a size limit
func (pt PolicyType) IsSizeLimited() bool {
	return (pt == PolicyFileSize)
}
