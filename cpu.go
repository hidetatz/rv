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

type CPU struct {
	// program counter
	PC uint64
	// System Bus
	Bus *Bus

	Regs *Registers
}

func NewCPU() *CPU {
	return &CPU{
		PC:   0,
		Bus:  NewBus(),
		Regs: NewRegisters(),
	}
}

func (cpu *CPU) Run() {
	// TODO: eventually physical <-> virtual memory translation must take place here.

	fmt.Printf("[debug] PC: 0x%x\n", cpu.PC)

	// Fetch
	inst := uint32(cpu.Bus.Read(cpu.PC, 32))
	fmt.Printf("[debug] fetched: %b\n", inst)

	// Decodes the fetched 32-bit instruction.
	// Note that rd, funct3, rs1... does not always match the instruction format,
	// but they are decoded only by its location in the 32-bit.
	opcode := uint8(inst & 0x00_00_00_7f)
	rd := uint8(inst & 0x00_00_0F_80 >> 7)
	funct3 := uint8(inst & 0x00_00_70_00 >> 12)
	rs1 := uint8(inst & 0x00_0F_80_00 >> 15)
	rs2 := uint8(inst & 0x01_F0_00_00 >> 20)
	funct7 := uint8(inst & 0xFE_00_00_00 >> 25)

	// Exec
	switch opcode {
	case 0b011_0011:
		// R
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				// add
				// Arithmetic overflow should be ignored according to the RISC-V spec.
				// In Go, primitive + ignores the overflow.
				cpu.Regs.Write(rd, cpu.Regs.Read(rs1)+cpu.Regs.Read(rs2))
			case 0b010_0000:
				// sub
				cpu.Regs.Write(rd, cpu.Regs.Read(rs1)-cpu.Regs.Read(rs2))
			}
		case 0b001:
			// sll
			// In RV64I, only the low 6 bits of rs2 are used as the shift amount
			// This is the same as SRL and SRA
			shift := cpu.Regs.Read(rs2) & 0b111111
			cpu.Regs.Write(rd, cpu.Regs.Read(rs1)<<shift)
		case 0b010:
			// slt
			var v uint64 = 0
			// must compare as two's complement
			if int64(cpu.Regs.Read(rs1)) < int64(cpu.Regs.Read(rs2)) {
				v = 1
			}
			cpu.Regs.Write(rd, v)
		case 0b011:
			// sltu
			var v uint64 = 0
			// must compare as unsigned val
			if cpu.Regs.Read(rs1) < cpu.Regs.Read(rs2) {
				v = 1
			}
			cpu.Regs.Write(rd, v)
		case 0b100:
			// xor
			cpu.Regs.Write(rd, cpu.Regs.Read(rs1)^cpu.Regs.Read(rs2))
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				// srl
				shift := cpu.Regs.Read(rs2) & 0b111111
				cpu.Regs.Write(rd, cpu.Regs.Read(rs1)>>shift)
			case 0b010_0000:
				// sra
				shift := cpu.Regs.Read(rs2) & 0b111111
				cpu.Regs.Write(rd, uint64(int64(cpu.Regs.Read(rs1))>>shift))
			}
		case 0b110:
			// or
			cpu.Regs.Write(rd, cpu.Regs.Read(rs1)|cpu.Regs.Read(rs2))
		case 0b111:
			// and
			cpu.Regs.Write(rd, cpu.Regs.Read(rs1)&cpu.Regs.Read(rs2))
		}
	case 0b110_0111:
		// I
		// jalr
		fallthrough
	case 0b000_0011:
		// I
		fallthrough
	case 0b001_0011:
		// I
		fallthrough
	case 0b000_1111:
		// I
		fallthrough
	case 0b111_0011:
		// I
		fallthrough
	case 0b010_0011:
		// S
		fallthrough
	case 0b110_0011:
		// B
		fallthrough
	case 0b001_0111:
		// U
		// auipc
		fallthrough
	case pb011_0111:
		// U
		// lui
		fallthrough
	case 0b110_1111:
		// J
		// jal
		fallthrough
	default:
		// TODO: define exception and return it
		panic(fmt.Sprintf("unknown instruction: %b", inst))
	}

	cpu.PC += 4
}
