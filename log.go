package main

import "fmt"

func Debug(format string, a ...any) {
	if debug {
		fmt.Printf("[debug] %s\n", fmt.Sprintf(format, a...))
	}
}
