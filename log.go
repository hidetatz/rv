package main

import "fmt"

// Debug writes debug log
func Debug(format string, a ...any) {
	if debug {
		fmt.Printf("[debug] %s\n", fmt.Sprintf(format, a...))
	}
}
