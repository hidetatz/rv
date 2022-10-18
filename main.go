package main

import (
	"debug/elf"
	"flag"
	"fmt"
	"os"
)

var dbg bool

type RV struct {
	cpu *CPU
	// tohost is an special address which shows a message from program to the host.
	// For now, tohost is used to terminate the execution of riscv-tests program.
	// https://riscv.org/wp-content/uploads/2015/01/riscv-testing-frameworks-bootcamp-jan2015.pdf
	tohost uint64
}

func (r *RV) Start() error {
	for {
		r.cpu.tick()

		if r.tohost == 0 {
			continue
		}

		if code := r.cpu.ram.Read(r.tohost, word); code != 0 {
			if code == 1 {
				fmt.Println("done")
				return nil
			}

			return fmt.Errorf("terminated, the tohost code is not success but %v", code)
		}
	}
}

func debug(format string, a ...any) {
	if dbg {
		fmt.Printf("[debug] %s\n", fmt.Sprintf(format, a...))
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		program = flag.String("p", "", "ELF program to run")
		d       = flag.Bool("d", false, "print out debug log if specified")
	)

	flag.Parse()

	dbg = *d

	file := *program
	if file == "" {
		return fmt.Errorf("program must be passed with -p option")
	}

	cpu, err := initCPU(file)
	if err != nil {
		return fmt.Errorf("initialize emulator: %w", err)
	}

	if err := cpu.Start(); err != nil {
		return fmt.Errorf("run program: %w", err)
	}

	return nil
}

func initCPU(filename string) (*RV, error) {
	of, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open elf file: %w", err)
	}

	// load elf file
	f, err := elf.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open elf file: %w", err)
	}

	if f.Data != elf.ELFDATA2LSB {
		return nil, fmt.Errorf("elf must be little endian")
	}

	if f.Type != elf.ET_EXEC {
		return nil, fmt.Errorf("elf type must be ET_EXEC")
	}

	if f.Machine != elf.EM_RISCV {
		return nil, fmt.Errorf("elf machine must be RISCV")
	}

	cpu := NewCPU()

	for _, p := range f.Progs {
		if p.Type != elf.PT_LOAD {
			continue
		}

		for i := 0; i < int(p.Filesz); i++ {
			addr := p.Vaddr + uint64(i)
			val := make([]byte, byt)
			_, err := of.ReadAt(val, int64(p.Off)+int64(i))
			if err != nil {
				return nil, fmt.Errorf("read program header")
			}
			cpu.ram.Write(addr, uint64(val[0]), byt)
		}
	}
	cpu.pc = f.Entry

	rv := &RV{cpu: cpu, tohost: 0x80001000} // TODO: find tohost from sections

	return rv, nil
}
