package main

// bits returns val[hi:lo].
func bits(val uint64, hi, lo int) uint64 {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

// Decode returns the format of the instruction.
func (cpu *CPU) Decode(inst uint64) InstructionCode {
	opcode := bits(inst, 0, 6)
	funct7 := bits(inst, 31, 25)
	funct3 := bits(inst, 14, 12)

	switch opcode {
	// RV64I
	case 0b011_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				return ADD
			case 0b010_0000:
				return SUB
			default:
				return _INVALID
			}
		case 0b001:
			return SLL
		case 0b010:
			return SLT
		case 0b011:
			return SLTU
		case 0b100:
			return XOR
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				return SRL
			case 0b010_0000:
				return SRA
			default:
				return _INVALID
			}
		case 0b110:
			return OR
		case 0b111:
			return AND
		default:
			return _INVALID
		}
	case 0b110_0111:
		return JALR
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			return LB
		case 0b001:
			return LH
		case 0b010:
			return LW
		case 0b100:
			return LBU
		case 0b101:
			return LHU
		default:
			return _INVALID
		}
	case 0b001_0011:
		switch funct3 {
		case 0b000:
			return ADDI
		case 0b010:
			return SLTI
		case 0b011:
			return SLTIU
		case 0b100:
			return XORI
		case 0b110:
			return ORI
		case 0b111:
			return ANDI
		case 0b001:
			return SLLI
		case 0b101:
			imm := bits(inst, 31, 20)
			switch imm >> 5 {
			case 0b000_0000:
				return SRLI
			case 0b010_0000:
				return SRAI
			default:
				return _INVALID
			}
		default:
			return _INVALID
		}
	case 0b000_1111:
		switch funct3 {
		case 0b000:
			return FENCE
		case 0b001:
			return FENCE_I
		default:
			return _INVALID
		}
	case 0b111_0011:
		switch funct3 {
		case 0b000:
			imm := bits(inst, 31, 20)
			switch imm {
			case 0b0:
				return ECALL
			case 0b1:
				return EBREAK
			default:
				return _INVALID
			}
		case 0b001:
			return CSRRW
		case 0b010:
			return CSRRS
		case 0b011:
			return CSRRC
		case 0b101:
			return CSRR2WI
		case 0b110:
			return CSRRSI
		case 0b111:
			return CSRRCI
		default:
			return _INVALID
		}
	case 0b001_0111:
		return AUIPC
	case 0b011_0111:
		return LUI
	case 0b110_1111:
		return JAL
	default:
		return _INVALID
	}
}
