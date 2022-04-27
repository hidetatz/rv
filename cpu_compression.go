package main

// IsCompressed returns if the instruction is compressed 16-bit one.
func (cpu *CPU) IsCompressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}

// Decompress extracts the given 16-bit instruction to 32-bit one.
func (cpu *CPU) DecodeCompressed(compressed uint64) InstructionCode {
	bs := func(hi, lo int) uint64 {
		return bits(compressed, hi, lo)
	}

	op := bs(1, 0)

	switch op {
	case 0b00:
		rdRs2 := bs(4, 2)
		if rdRs2 == 0b0 {
			return _INVALID
		}

		funct3 := bs(15, 13)
		switch funct3 {
		case 0b000:
			return C_ADDI4SPN
		case 0b001:
			return C_FLD
		case 0b010:
			return C_LW
		case 0b011:
			return C_LD
		case 0b101:
			return C_FSD
		case 0b110:
			return C_SW
		case 0b111:
			return C_SD
		}
	case 0b01:
		funct3 := bs(15, 13)
		switch funct3 {
		case 0b000:
			rdRs1 := bs(11, 7)
			switch rdRs1 {
			case 0b0:
				return _INVALID
			case 0b11:
				return C_ADDI16SP
			default:
				return C_LUI
			}
		case 0b001:
			return C_ADDIW
		case 0b010:
			return C_LI
		case 0b011:
			rdRs1 := bs(11, 7)
			switch rdRs1 {
			case 0b0:
				return C_NOP
			default:
				return C_ADDI
			}
		case 0b100:
			switch bs(11, 10) {
			case 0b00:
				return C_SRLI
			case 0b01:
				return C_SRAI
			case 0b10:
				return C_ANDI
			case 0b11:
				switch bit(compressed, 12) {
				case 0b0:
					switch bs(6, 5) {
					case 0b00:
						return C_SUB
					case 0b01:
						return C_XOR
					case 0b10:
						return C_OR
					case 0b11:
						return C_AND
					}
				case 0b1:
					switch bs(6, 5) {
					case 0b00:
						return C_SUBW
					case 0b01:
						return C_ADDW
					}
				}
			}
		case 0b101:
			return C_J
		case 0b110:
			return C_BEQZ
		case 0b111:
			return C_BNEZ
		}
	case 0b10:
		funct3 := bits(compressed, 15, 13)
		switch funct3 {
		case 0b000:
			return C_SLLI
		case 0b001:
			return C_FLDSP
		case 0b010:
			return C_LWSP
		case 0b011:
			return C_LDSP
		case 0b100:
			switch bit(compressed, 12) {
			case 0b0:
				switch bits(compressed, 6, 2) {
				case 0b0:
					return C_JR
				default:
					return C_MV
				}
			case 0b1:
				switch bits(compressed, 11, 7) {
				case 0b0:
					return C_EBREAK
				default:
					switch bits(compressed, 6, 2) {
					case 0b0:
						return C_JALR
					default:
						return C_ADD
					}
				}
			}
		case 0b101:
			return C_FSDSP
		case 0b110:
			return C_SWSP
		case 0b111:
			return C_SDSP
		}
	}

	return _INVALID
}
