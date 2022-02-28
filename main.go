package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var (
		program = flag.String("p", "", "ELF program to run")
	)

	flag.Parse()

	file := *program
	if file == "" {
		fmt.Fprintf(os.Stderr, "Err: program must be passed with -p option\n")
		os.Exit(1)
	}

	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: open program: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	buff, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: read program: %v\n", err)
		os.Exit(1)
	}

	emulator := New(buff)
	emulator.Start()
}
