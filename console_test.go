package console

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
)

func Test(t *testing.T) {
	running()
}

func running() {
	defer New(&Options{
		Info: true, Debug: true, Warning: true, Error: true, Print: true,
		LogFileSizeMB: 100,
		MaxBackups:    3,
		Filename:      "log/execution.log",
	}).Wait()

	txt := "error message."

	// INFO
	INFO(txt)

	// DEBUG
	DEBUG(txt)

	// WARN
	warn := WARN(txt)
	fmt.Printf("WARN: %s\n", warn)

	// ERROR
	err := ERROR(errors.New(txt))
	fmt.Printf("ERROR: %s\n", err)
}
