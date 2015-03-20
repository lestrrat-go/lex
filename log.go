package lex

import (
	"log"
	"os"
	"strconv"
)

var debug = false
var logger = log.New(os.Stderr, "", 0)

func init() {
	v, err := strconv.ParseBool(os.Getenv("LEX_DEBUG"))
	if err == nil {
		debug = v
	}
}

// Trace outputs log if debug is enabled
func Trace(format string, args ...interface{}) {
	if !debug {
		return
	}
	logger.Printf(format, args...)
}

// Mark marks the begin/end of a function, and also indents the log output
// accordingly
func Mark(format string, args ...interface{}) func() {
	if !debug {
		return func() {}
	}

	Trace("START "+format, args...)
	old := logger.Prefix()
	logger.SetPrefix(old + "|    ")
	return func() {
		logger.SetPrefix(old)
		Trace("END "+format, args...)
	}
}
