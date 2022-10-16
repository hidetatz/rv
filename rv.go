package main

import "fmt"

// RV is a central RISC-V emulator.
type RV struct {
	cpu *CPU
	// tohost is an special address which shows a message from program to the host.
	// For now, tohost is used to terminate the execution of riscv-tests program.
	// https://riscv.org/wp-content/uploads/2015/01/riscv-testing-frameworks-bootcamp-jan2015.pdf
	tohost uint64
}

// New initializes and returns the RISC-V emulator rv.
// The argument program must be the ELF binary which is built for RISC-V architecture.
func New(prog []byte) (*RV, error) {
	elf, err := LoadELF(prog)
	if err != nil {
		return nil, fmt.Errorf("Load ELF file: %w", err)
	}

	if elf.Header.Data != 1 { // Little endian
		return nil, fmt.Errorf("ELF data is not ET_EXEC little endian but %d. Cannot execute", elf.Header.Data)
	}

	if elf.Header.Type != 2 { // ET_EXEC
		return nil, fmt.Errorf("ELF type is not ET_EXEC but %d. Cannot execute", elf.Header.Type)
	}

	if elf.Header.Machine != 0xf3 { // RISC-V
		return nil, fmt.Errorf("ELF machine is not RISC-V but %d. Cannot execute", elf.Header.Machine)
	}

	if elf.Header.PhNum == 0 { // assert just in case
		return nil, fmt.Errorf("ELF contains no program headers. Cannot execute")
	}

	xlen := xlen32
	if elf.Header.Class == 2 {
		xlen = xlen64
	}

	cpu := NewCPU(xlen)

	for _, p := range elf.Programs {
		if p.Type != 1 { // PT_LOAD
			continue
		}

		// write to memory
		for i := 0; i < int(p.Filesz); i++ {
			addr := p.VAddr + uint64(i)
			val := uint64(prog[int(p.Offset)+i])
			cpu.Write(addr, val, byt)
		}
	}
	cpu.PC = elf.Header.Entry

	rv := &RV{cpu: cpu, tohost: elf.ToHost}

	Debug("Load ELF succeeded: PC: %x, ToHost: %x", cpu.PC, rv.tohost)

	return rv, nil
}

// Start starts the emulator fetch-decode-exec cycle.
// It runs a loop until a Fatal level trap occurs.
// It optionally can stop running if the given binary contains .tohost address,
// which is a part of RISC-V specification.
func (r *RV) Start() error {
	for {
		trap := r.cpu.Run()

		// For now, only handle Fatal trap to terminate the program execution.
		if trap == TrapFatal {
			return fmt.Errorf("Fatal trap is returned!")
		}

		if r.tohost == 0 {
			continue
		}

		if code := r.cpu.MMU.ram.Read(r.tohost, word); code != 0 {
			if code == 1 {
				return nil
			}

			return fmt.Errorf("terminated, the tohost code is not success but %v", code)
		}
	}
}

// Debug writes debug log
func Debug(format string, a ...any) {
	if debug {
		fmt.Printf("[debug] %s\n", fmt.Sprintf(format, a...))
	}
}
