package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var debug bool

func main() {
	var (
		program = flag.String("p", "", "ELF program to run")
		dbg     = flag.Bool("d", false, "print out debug log if specified")
	)

	flag.Parse()

	debug = *dbg

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

	emulator, err := New(buff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: initialize program: %v\n", err)
		os.Exit(1)
	}

	emulator.Start()
}
