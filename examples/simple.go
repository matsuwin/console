package main

import (
	"fmt"
	"github.com/pkg/errors"
	"console"
)

func main() {
	running()
}

func running() {
	defer console.New(&console.Options{
		Info: true, Debug: true, Warning: true, Error: true, Print: true,
		LogFileSizeMB: 1024,
		MaxBackups:    10,
		Filename:      "log/execution.log",
	}).Wait()

	txt := "error message."

	// INFO
	console.INFO(txt)

	// DEBUG
	console.DEBUG(txt)

	// WARN
	warn := console.WARN(txt)
	fmt.Printf("WARN: %s\n", warn)

	// ERROR
	err := console.ERROR(errors.New(txt))
	fmt.Printf("ERROR: %s\n", err)
}
