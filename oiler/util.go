package main

import (
	"fmt"
	"os"
)

func Print(format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", params...)
}

func Debug(format string, params ...interface{}) {
	Print("DEBUG: "+format, params...)
}

func Info(format string, params ...interface{}) {
	Print("INFO:  "+format, params...)
}

func Error(err error) {
	Print("ERROR: %s", err)
}

func Fatal(err error, exitcode int) {
	Print("FATAL: %s", err)
	os.Exit(exitcode)
}
