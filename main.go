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

// Instruction is a general-purpose 32-bit instruction representation.
// This is to hold the decoded raw instruction binary string.
// Some fields such as Rd, Funct3, Rs1... are called other name like imm or shamt in some instructions,
// but the most common and widely used names are chosen.
type Instruction struct {
	Raw    uint32 // raw instruction
	Opcode uint8  // 7-bit (from 0 to 6).
	Rd     uint8  // 5-bit (from 7 to 11).
	Funct3 uint8  // 3-bit (from 12 to 14).
	Rs1    uint8  // 5-bit (from 15 to 19).
	Rs2    uint8  // 5-bit (from 20 to 24).
	Funct7 uint8  // 7-bit (from 25 to 31).
}

// Decode decodes the given 32-bit instruction.
// Note that rd, funct3, rs1... does not always match the instruction format,
// but they are just decoded by the location in the 32-bit.
// With that consideration, the raw instruction binary is also stored in the returned struct.
func (cpu *CPU) Decode(ins uint32) uint64 {
	opcode := ins & 0x00_00_00_7f

	switch opcode {
	// R
	case 0b011_0011:
		funct3 := ins & 0x00_00_38_00 >> 12
		switch funct3 {
		case 0b000:
			funct7 := ins & 0xFE_00_00_00 >> 25
			switch funct7 {
			// add
			case 0b000_0000:
			// sub
			case 0b010_0000:
			}
		// sll
		case 0b001:
		// slt
		case 0b010:
		// sltu
		case 0b011:
		// xor
		case 0b100:
		case 0b101:
			funct7 := ins & 0xFE_00_00_00 >> 25
			switch funct7 {
			// srl
			case 0b000_0000:
			// sra
			case 0b010_0000:
			}
		// or
		case 0b110:
		// and
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
