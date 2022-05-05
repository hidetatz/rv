package main

// Decode decodes the given binary and returns which instruction
// the binary represents.
// It cannot decode the compressed instruction. For that purpose,
// DecodeCompressed must be used.
func (cpu *CPU) Decode(inst uint64) InstructionCode {
	opcode := bits(inst, 6, 0)
	funct7 := bits(inst, 31, 25)
	funct3 := bits(inst, 14, 12)

	switch opcode {
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			return LB
		case 0b001:
			return LH
		case 0b010:
			return LW
		case 0b011:
			return LD
		case 0b100:
			return LBU
		case 0b101:
			return LHU
		case 0b110:
			return LWU
		}
	case 0b000_1111:
		switch funct3 {
		case 0b000:
			return FENCE
		case 0b001:
			return FENCE_I
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
			}
		}
	case 0b001_0111:
		return AUIPC
	case 0b001_1011:
		switch funct3 {
		case 0b000:
			return ADDIW
		default:
			switch funct7 {
			case 0b000_0000:
				switch funct3 {
				case 0b001:
					return SLLIW
				case 0b101:
					return SRLIW
				}
			case 0b010_0000:
				return SRAIW
			}
		}
	case 0b010_0011:
		switch funct3 {
		case 0b000:
			return SB
		case 0b001:
			return SH
		case 0b010:
			return SW
		case 0b011:
			return SD
		}
	case 0b010_1111:
		switch funct3 {
		case 0b010:
			switch bits(inst, 31, 27) {
			case 0b0_0010:
				return LR_W
			case 0b0_0011:
				return SC_W
			case 0b0_0001:
				return AMOSWAP_W
			case 0b0_0000:
				return AMOADD_W
			case 0b0_0100:
				return AMOXOR_W
			case 0b0_1100:
				return AMOAND_W
			case 0b0_1000:
				return AMOOR_W
			case 0b1_0000:
				return AMOMIN_W
			case 0b1_0100:
				return AMOMAX_W
			case 0b1_1000:
				return AMOMINU_W
			case 0b1_1100:
				return AMOMAXU_W
			}
		case 0b011:
			switch bits(inst, 31, 27) {
			case 0b0_0010:
				return LR_D
			case 0b0_0011:
				return SC_D
			case 0b0_0001:
				return AMOSWAP_D
			case 0b0_0000:
				return AMOADD_D
			case 0b0_0100:
				return AMOXOR_D
			case 0b0_1100:
				return AMOAND_D
			case 0b0_1000:
				return AMOOR_D
			case 0b1_0000:
				return AMOMIN_D
			case 0b1_0100:
				return AMOMAX_D
			case 0b1_1000:
				return AMOMINU_D
			case 0b1_1100:
				return AMOMAXU_D
			}
		}
	case 0b011_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				return ADD
			case 0b000_0001:
				return MUL
			case 0b010_0000:
				return SUB
			}
		case 0b001:
			switch funct7 {
			case 0b000_0001:
				return MULH
			default:
				return SLL
			}
		case 0b010:
			switch funct7 {
			case 0b000_0001:
				return MULHSU
			default:
				return SLT
			}
		case 0b011:
			switch funct7 {
			case 0b000_0001:
				return MULHU
			default:
				return SLTU
			}
		case 0b100:
			switch funct7 {
			case 0b000_0001:
				return DIV
			default:
				return XOR
			}
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				return SRL
			case 0b000_0001:
				return DIVU
			case 0b010_0000:
				return SRA
			}
		case 0b110:
			switch funct7 {
			case 0b000_0001:
				return REM
			default:
				return OR
			}
		case 0b111:
			switch funct7 {
			case 0b000_0001:
				return REMU
			default:
				return AND
			}
		}
	case 0b011_0111:
		return LUI
	case 0b011_1011:
		switch funct7 {
		case 0b000_0000:
			switch funct3 {
			case 0b000:
				return ADDW
			case 0b001:
				return SLLW
			case 0b101:
				return SRLW
			}
		case 0b000_0001:
			switch funct3 {
			case 0b000:
				return MULW
			case 0b100:
				return DIVW
			case 0b101:
				return DIVUW
			case 0b110:
				return REMW
			case 0b111:
				return REMUW
			}
		case 0b010_0000:
			switch funct3 {
			case 0b000:
				return SUBW
			case 0b101:
				return SRAW
			}
		}
	case 0b110_0011:
		switch funct3 {
		case 0b000:
			return BEQ
		case 0b001:
			return BNE
		case 0b100:
			return BLT
		case 0b101:
			return BGE
		case 0b110:
			return BLTU
		case 0b111:
			return BGEU
		}
	case 0b110_0111:
		return JALR
	case 0b110_1111:
		return JAL
	case 0b111_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_1000:
				switch bits(inst, 24, 20) {
				case 0b0_0010:
					return SRET
				case 0b0_0101:
					return WFI
				}
			case 0b001_1000:
				return MRET
			case 0b000_1001:
				return SFENCE_VMA
			case 0b000_0000:
				imm := bits(inst, 24, 20)
				switch imm {
				case 0b00:
					return ECALL
				case 0b01:
					return EBREAK
				case 0b10:
					return URET
				}
			}
		case 0b001:
			return CSRRW
		case 0b010:
			return CSRRS
		case 0b011:
			return CSRRC
		case 0b101:
			return CSRRWI
		case 0b110:
			return CSRRSI
		case 0b111:
			return CSRRCI
		}
	}

	return _INVALID
}
