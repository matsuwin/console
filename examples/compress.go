package main

import "console"

func main() {
	defer console.New(&console.Options{
		Info: true, Debug: true, Warning: true, Error: true,
		Print:         false,
		LogFileSizeMB: 4,
		MaxBackups:    2,
		Filename:      "log/execution.log",
	}).Wait()

	txt := `###########################################################################################################`

	for i := 0; i < 100000; i++ {
		console.INFO(txt)
	}
}
