package main

// CompressedInstructionFormat is a format for RV32C/RV64C.
type CompressedInstructionFormat uint8

const (
	CompressedInstructionFormatInvalid CompressedInstructionFormat = iota
	CompressedInstructionFormatCR
	CompressedInstructionFormatCI
	CompressedInstructionFormatCSS
	CompressedInstructionFormatCIW
	CompressedInstructionFormatCL
	CompressedInstructionFormatCS
	CompressedInstructionFormatCB
	CompressedInstructionFormatCJ
)

// Compressed returns if the instruction is compressed 16-bit one.
func (cpu *CPU) Compressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, is it 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}

// Uncompress extracts the given 16-bit instruction to 32-bit one.
func (cpu *CPU) Uncompress(compressed uint64) (uint64, Exception) {
	compressedInstructionFormat := cpu.DecodeCompressedInstructionFormat(compressed)
	if compressedInstructionFormat == CompressedInstructionFormatInvalid {
		return 0, ExcpIllegalInstruction
	}

	switch compressedInstructionFormat {
	case CompressedInstructionFormatCR:
		return cpu.UncompressCR(
			bits(compressed, 1, 0),
			bits(compressed, 6, 2),
			bits(compressed, 11, 7),
			bits(compressed, 15, 12),
		)
	case CompressedInstructionFormatCI:
	case CompressedInstructionFormatCSS:
	case CompressedInstructionFormatCIW:
	case CompressedInstructionFormatCL:
	case CompressedInstructionFormatCS:
	case CompressedInstructionFormatCB:
	case CompressedInstructionFormatCJ:
	}

	return 0, ExcpIllegalInstruction
}

func (cpu *CPU) UncompressCR(op, rs2, rdrs1, funct4 uint64) (uint64, Exception) {
	switch op {
	}

	return 0, ExcpNone
}

// DecodeCompressedInstructionFormat decodes the given compressed instruction and returns its format.
func (cpu *CPU) DecodeCompressedInstructionFormat(compressed uint64) CompressedInstructionFormat {
	opcode := compressed & 0b11
	funct3 := (compressed >> 13) & 0b111
	switch opcode {
	case 0b00:
		switch funct3 {
		case 0b000:
			return CompressedInstructionFormatCIW
		case 0b001, 0b010, 0b011, 0b101, 0b110, 0b111:
			return CompressedInstructionFormatCL
		default:
			return CompressedInstructionFormatInvalid
		}
	case 0b01:
		switch funct3 {
		case 0b000, 0b010, 0b011:
			return CompressedInstructionFormatCI
		case 0b100:
			f := (compressed >> 10) & 0b11
			switch f {
			case 0b00, 0b01, 0b10:
				return CompressedInstructionFormatCI
			case 0b11:
				return CompressedInstructionFormatCR
			default:
				return CompressedInstructionFormatInvalid
			}
		case 0b001, 0b101:
			return CompressedInstructionFormatCJ
		case 0b110, 0b111:
			return CompressedInstructionFormatCB
		default:
			return CompressedInstructionFormatInvalid
		}
	case 0b10:
		switch funct3 {
		case 0b000:
			return CompressedInstructionFormatCI
		case 0b100:
			f1 := (compressed >> 12) & 0b1
			f2 := (compressed >> 2) & 0b1_1111
			f3 := (compressed >> 7) & 0b1_1111
			switch f1 {
			case 0b0:
				if f2 == 0 {
					return CompressedInstructionFormatCJ
				} else {
					return CompressedInstructionFormatCR
				}
			case 0b1:
				if f2 == 0 && f3 == 0 {
					return CompressedInstructionFormatCI
				} else if f2 == 0 {
					return CompressedInstructionFormatCJ
				} else {
					return CompressedInstructionFormatCR
				}
			}
		case 0b001, 0b010, 0b011, 0b101, 0b110, 0b111:
			return CompressedInstructionFormatCSS
		default:
			return CompressedInstructionFormatInvalid
		}
	default:
		return CompressedInstructionFormatInvalid
	}

	return CompressedInstructionFormatInvalid
}
