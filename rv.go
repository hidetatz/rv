package main

import "fmt"

type RV struct {
	cpu *CPU
	// tohost is an special address which shows a message from program to the host.
	// For now, tohost is used to terminate the execution of riscv-tests program.
	// https://riscv.org/wp-content/uploads/2015/01/riscv-testing-frameworks-bootcamp-jan2015.pdf
	tohost uint64
}

func New(prog []byte) (*RV, error) {
	cpu := NewCPU()

	elf, err := LoadELF(prog)
	if err != nil {
		return nil, fmt.Errorf("Load ELF file: %w", err)
	}

	if elf.Header.Class != 2 { // 64-bit
		return nil, fmt.Errorf("ELF class is not 64-bit but %d. Cannot execute", elf.Header.Class)
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
		return nil, fmt.Errorf("ELF contains no program headers. Cannot execute", elf.Header.Machine)
	}

	for _, p := range elf.Programs {
		if p.Type != 1 { // PT_LOAD
			continue
		}

		// write to memory
		for i := 0; i < int(p.Filesz); i++ {
			addr := p.VAddr + uint64(i)
			val := uint64(prog[int(p.Offset)+i])
			cpu.Bus.Write(addr, val, Byte)
		}
	}
	cpu.PC = elf.Header.Entry

	rv := &RV{cpu: cpu, tohost: elf.ToHost}

	Debug("Load ELF succeeded: PC: %x, ToHost: %x", cpu.PC, rv.tohost)

	return rv, nil
}

func (r *RV) Start() {
	for {
		trap := r.cpu.Run()

		// For now, only handle Fatal trap to terminate the program execution.
		if trap == TrapFatal {
			fmt.Println("Fatal trap is returned!")
			return
		}

		if r.tohost == 0 {
			continue
		}

		if code := r.cpu.Bus.Read(r.tohost, Word); code != 0 {
			if code == 1 {
				fmt.Println("Successfully done")
				return
			}
			fmt.Printf("fail: %v\n", code)
			return
		}
	}
}
