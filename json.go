package logger

import (
	"encoding/json"
	"log"
)

// JSONFormatter for logger
type JSONFormatter struct {
	name string
}

// JSONFormatter creates a new formatter for logger
func Json() *JSONFormatter {
	return &JSONFormatter{name: "json"}
}

// Format implements the EventFormatter interface
func (jf *JSONFormatter) Format(em EventMsg) (msg string, err error) {
	bMsg, jErr := json.Marshal(em)
	if jErr != nil {
		log.Printf("Json error: %s (%+v)\n", jErr, em)
		return "", jErr
	}
	return string(bMsg), nil
}
