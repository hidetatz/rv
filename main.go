package main

import (
	"fmt"
)

type Bus struct {
	Memory *Memory
	// TODO: Add some more peripheral devices, such as UART.
}

func NewBus() *Bus {
	return &Bus{
		Memory: NewMemory(),
	}
}

func (bus *Bus) Read(addr uint64, size uint8) uint64 {
	return bus.Memory.Read(addr, size)
}

type Registers struct {
	// every register size is 32bit
	Regs [32]uint8
}

func NewRegisters() *Registers {
	return &Registers{Regs: [32]uint8{}}
}

type CPU struct {
	// program counter
	PC uint64
	// System Bus
	Bus *Bus

	Registers *Registers
}

func NewCPU() *CPU {
	return &CPU{
		PC:        0,
		Bus:       NewBus(),
		Registers: NewRegisters(),
	}
}

// TODO: return Exception.
func (cpu *CPU) Fetch(size uint8) uint64 {
	if size != 32 && size != 16 {
		panic(fmt.Sprintf("invalid instruction size requested: %d", size))
	}

	// TODO: eventually physical <-> virtual memory translation must take place here.

	return cpu.Bus.Read(cpu.PC, size)
}

// Exec executes the given 32-bit instruction.
// TODO: support 16-bit instructions
func (cpu *CPU) Exec(ins uint32) uint64 {
	opcode := ins & 0x00_00_00_7f

	switch opcode {
	// R
	case 0b011_0011:
		funct3 := ins & 0x00_00_38_00 >> 12
		switch funct3 {
		case 0b000:
			funct7 := ins & 0xFE_00_00_00 >> 25
			switch funct7 {
			case 0b000_0000:
			case 0b010_0000:
			}
		case 0b001:
		case 0b010:
		case 0b011:
		case 0b100:
		case 0b101:
			funct7 := ins & 0xFE_00_00_00 >> 25
			switch funct7 {
			case 0b000_0000:
			case 0b010_0000:
			}
		case 0b110:
		case 0b111:
		}
	// I
	case 0b110_0111, 0b000_0011, 0b001_0011, 0b000_1111, 0b111_0011:
		fallthrough
	// S
	case 0b010_0011:
		fallthrough
	// B
	case 0b110_0011:
		fallthrough
	// U
	case 0b011_0111, 0b001_0111:
		fallthrough
	// J
	case 0b110_1111:
		fallthrough
	default:
		panic(fmt.Sprintf("unimplemented instruction: %b", ins))
	}
}

func main() {
}
