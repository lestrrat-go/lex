package lex

import (
	"log"
	"os"
	"strconv"
)

var Debug = false
var Logger = log.New(os.Stderr, "", 0)

func init() {
	v, err := strconv.ParseBool(os.Getenv("LEX_DEBUG"))
	if err == nil {
		Debug = v
	}
}

// Trace outputs log if debug is enabled
func Trace(format string, args ...interface{}) {
	if !Debug {
		return
	}
	Logger.Printf(format, args...)
}

// Mark marks the begin/end of a function, and also indents the log output
// accordingly
func Mark(format string, args ...interface{}) func() {
	if !Debug {
		return func() {}
	}

	Trace("START "+format, args...)
	old := Logger.Prefix()
	Logger.SetPrefix(old + "|    ")
	return func() {
		Logger.SetPrefix(old)
		Trace("END "+format, args...)
	}
}
