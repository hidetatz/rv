package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var debug bool

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		program = flag.String("p", "", "ELF program to run")
		dbg     = flag.Bool("d", false, "print out debug log if specified")
	)

	flag.Parse()

	debug = *dbg

	file := *program
	if file == "" {
		return fmt.Errorf("program must be passed with -p option")
	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open program: %w", err)
	}
	defer f.Close()

	buff, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read program: %w", err)
	}

	emulator, err := New(buff)
	if err != nil {
		return fmt.Errorf("initialize emulator: %w", err)
	}

	if err := emulator.Start(); err != nil {
		return fmt.Errorf("run program: %w", err)
	}

	return nil
}
