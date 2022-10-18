package main

import (
	"flag"
	"fmt"
	"io"
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

func New(prog []byte) (*RV, error) {
	elf, err := LoadELF(prog)
	if err != nil {
		return nil, fmt.Errorf("load ELF: %w", err)
	}

	if elf.Header.Data != 1 { // Little endian
		return nil, fmt.Errorf("elf data is not ET_EXEC little endian but %d", elf.Header.Data)
	}

	if elf.Header.Type != 2 { // ET_EXEC
		return nil, fmt.Errorf("elf type is not ET_EXEC but %d", elf.Header.Type)
	}

	if elf.Header.Machine != 0xf3 { // RISC-V
		return nil, fmt.Errorf("elf machine is not RISC-V but %d", elf.Header.Machine)
	}

	if elf.Header.PhNum == 0 { // assert just in case
		return nil, fmt.Errorf("elf contains no program headers")
	}

	if elf.Header.Class != 2 {
		return nil, fmt.Errorf("only 64 bit is supported")
	}

	cpu := NewCPU()

	for _, p := range elf.Programs {
		if p.Type != 1 { // PT_LOAD
			continue
		}

		// write to memory
		for i := 0; i < int(p.Filesz); i++ {
			addr := p.VAddr + uint64(i)
			val := uint64(prog[int(p.Offset)+i])
			cpu.write(addr, val, byt)
		}
	}
	cpu.pc = elf.Header.Entry

	rv := &RV{cpu: cpu, tohost: elf.ToHost}

	return rv, nil
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
