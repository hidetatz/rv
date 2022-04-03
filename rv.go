package main

import "fmt"

type RV struct {
	cpu *CPU
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
			cpu.Bus.Write(p.VAddr+uint64(i), uint64(prog[int(p.Offset)+i]), Byte)
		}
	}

	cpu.PC = elf.Header.Entry

	return &RV{cpu: cpu}, nil
}

func (r *RV) Start() {
	for {
		r.cpu.Run()
	}
}

func (r *RV) StartForTest(count int) {
	for i := 0; i < count; i++ {
		r.cpu.Run()
	}
}
