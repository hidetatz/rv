package rv

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

// Decode32 decodes the given 32-bit instruction.
// Note that rd, funct3, rs1... does not always match the instruction format,
// but they are just decoded by the location in the 32-bit.
// With that consideration, the raw instruction binary is also stored in the returned struct.
// TODO: return Exception.
func (cpu *CPU) Decode32(inst uint32) *Instruction {
	ins := &Instruction{
		Raw:    inst,
		Opcode: uint8(inst & 0x00_00_00_7f),
		Rd:     uint8(inst & 0x00_00_0F_80 >> 7),
		Funct3: uint8(inst & 0x00_00_70_00 >> 12),
		Rs1:    uint8(inst & 0x00_0F_80_00 >> 15),
		Rs2:    uint8(inst & 0x01_F0_00_00 >> 20),
		Funct7: uint8(inst & 0xFE_00_00_00 >> 25),
	}
	// still need to fill Op

	switch ins.Opcode {
	case 0b011_0011:
		switch ins.Funct3 {
		case 0b000:
			switch ins.Funct7 {
			case 0b000_0000:
				ins.Op = InsAdd
			case 0b010_0000:
				ins.Op = InsSub
			}
		case 0b001:
			ins.Op = InsSLL
		case 0b010:
			ins.Op = InsSLT
		case 0b011:
			ins.Op = InsSLTU
		case 0b100:
			ins.Op = InsXOR
		case 0b101:
			switch ins.Funct7 {
			case 0b000_0000:
				ins.Op = InsSRL
			case 0b010_0000:
				ins.Op = InsSRA
			}
		case 0b110:
			ins.Op = InsOr
		case 0b111:
			ins.Op = InsAnd
		}
	case 0b110_0111, 0b000_0011, 0b001_0011, 0b000_1111, 0b111_0011:
		fallthrough
	case 0b010_0011:
		fallthrough
	case 0b110_0011:
		fallthrough
	case 0b011_0111, 0b001_0111:
		fallthrough
	case 0b110_1111:
		fallthrough
	default:
		// TODO: define exception and return it
		panic(fmt.Sprintf("unknown instruction: %b", ins))
	}

	return ins
}

func (cpu *CPU) Exec(inst *Instruction) {
	ins := &Instruction{
		Raw:    inst,
		Opcode: uint8(inst & 0x00_00_00_7f),
		Rd:     uint8(inst & 0x00_00_0F_80 >> 7),
		Funct3: uint8(inst & 0x00_00_70_00 >> 12),
		Rs1:    uint8(inst & 0x00_0F_80_00 >> 15),
		Rs2:    uint8(inst & 0x01_F0_00_00 >> 20),
		Funct7: uint8(inst & 0xFE_00_00_00 >> 25),
	}
	// still fill Op

	switch ins.Opcode {
	case 0b011_0011:
		switch ins.Funct3 {
		case 0b000:
			switch ins.Funct7 {
			case 0b000_0000:
				ins.Op = InsAdd
			case 0b010_0000:
				ins.Op = InsSub
			}
		case 0b001:
			ins.Op = InsSLL
		case 0b010:
			ins.Op = InsSLT
		case 0b011:
			ins.Op = InsSLTU
		case 0b100:
			ins.Op = InsXOR
		case 0b101:
			switch ins.Funct7 {
			case 0b000_0000:
				ins.Op = InsSRL
			case 0b010_0000:
				ins.Op = InsSRA
			}
		case 0b110:
			ins.Op = InsOr
		case 0b111:
			ins.Op = InsAnd
		}
	case 0b110_0111, 0b000_0011, 0b001_0011, 0b000_1111, 0b111_0011:
		fallthrough
	case 0b010_0011:
		fallthrough
	case 0b110_0011:
		fallthrough
	case 0b011_0111, 0b001_0111:
		fallthrough
	case 0b110_1111:
		fallthrough
	default:
		// TODO: define exception and return it
		panic(fmt.Sprintf("unknown instruction: %b", ins))
	}

	return ins
}
