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
	inst := cpu.Fetch(32)
	dec := cpu.Decode32(inst)
	cpu.Exec(dec)
	cpu.PC += 4
}

// TODO: return Exception.
func (cpu *CPU) Fetch(size uint8) uint32 {
	if size != 32 {
		panic(fmt.Sprintf("invalid instruction size requested: %d", size))
	}

	// TODO: eventually physical <-> virtual memory translation must take place here.

	return uint32(cpu.Bus.Read(cpu.PC, size))
}

// Decode32 decodes the given 32-bit instruction.
// Note that rd, funct3, rs1... does not always match the instruction format,
// but they are just decoded by the location in the 32-bit.
// With that consideration, the raw instruction binary is also stored in the returned struct.
// TODO: return Exception.
func (cpu *CPU) Decode32(inst uint32) *Instruction {
	fmt.Printf("%b\n", inst)
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
	switch inst.Op {
	case InsAdd:
		// Arithmetic overflow should be ignored according to the RISC-V spec.
		// In Go, primitive + ignores the overflow.
		fmt.Println("yay")
		fmt.Println(cpu.Regs.Read(inst.Rs1) + cpu.Regs.Read(inst.Rs2))
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)+cpu.Regs.Read(inst.Rs2))
	case InsSub:
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)-cpu.Regs.Read(inst.Rs2))
	case InsSLL:
		// In RV64I, only the low 6 bits of rs2 are used as the shift amount
		// This is the same as SRL and SRA
		shift := cpu.Regs.Read(inst.Rs2) & 0b111111
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)<<shift)
	case InsSLT:
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.Regs.Read(inst.Rs1)) < int64(cpu.Regs.Read(inst.Rs2)) {
			v = 1
		}
		cpu.Regs.Write(inst.Rd, v)
	case InsSLTU:
		var v uint64 = 0
		// must compare as unsigned val
		if cpu.Regs.Read(inst.Rs1) < cpu.Regs.Read(inst.Rs2) {
			v = 1
		}
		cpu.Regs.Write(inst.Rd, v)
	case InsXOR:
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)^cpu.Regs.Read(inst.Rs2))
	case InsSRL:
		shift := cpu.Regs.Read(inst.Rs2) & 0b111111
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)>>shift)
	case InsSRA:
		shift := cpu.Regs.Read(inst.Rs2) & 0b111111
		cpu.Regs.Write(inst.Rd, uint64(int64(cpu.Regs.Read(inst.Rs1))>>shift))
	case InsOr:
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)|cpu.Regs.Read(inst.Rs2))
	case InsAnd:
		cpu.Regs.Write(inst.Rd, cpu.Regs.Read(inst.Rs1)&cpu.Regs.Read(inst.Rs2))
	}
}
