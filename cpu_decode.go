package main

type InstructionFormat uint8

const (
	InstructionFormatInvalid InstructionFormat = iota
	InstructionFormatR
	InstructionFormatI
	InstructionFormatS
	InstructionFormatB
	InstructionFormatU
	InstructionFormatJ
)

// DecodeInstructionFormat returns the format of the instruction.
func (cpu *CPU) DecodeInstructionFormat(inst uint64) InstructionFormat {
	opcode := inst & 0b111_1111
	switch opcode {
	case 0b011_0011:
		return InstructionFormatR
	case 0b110_0111, 0b000_0011, 0b001_0011, 0b000_1111, 0b111_0011:
		return InstructionFormatI
	case 0b010_0011:
		return InstructionFormatS
	case 0b110_0011:
		return InstructionFormatB
	case 0b011_0111, 0b001_0111:
		return InstructionFormatU
	case 0b110_1111:
		return InstructionFormatJ
	default:
		return InstructionFormatInvalid
	}
}
