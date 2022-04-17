package main

type Decoded struct {
	Code   InstructionCode
	Format InstructionFormat
}

// Decode returns the format of the instruction.
func (cpu *CPU) Decode(inst uint64) Decoded {
	opcode := bits(inst, 6, 0)
	funct7 := bits(inst, 31, 25)
	funct3 := bits(inst, 14, 12)

	Debug("opcode: %07b", opcode)
	Debug("funct7: %07b", funct7)
	Debug("funct3: %03b", funct3)

	switch opcode {
	// RV64I
	case 0b011_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				return Decoded{ADD, InstructionFormatR}
			case 0b010_0000:
				return Decoded{SUB, InstructionFormatR}
			}
		case 0b001:
			return Decoded{SLL, InstructionFormatR}
		case 0b010:
			return Decoded{SLT, InstructionFormatR}
		case 0b011:
			return Decoded{SLTU, InstructionFormatR}
		case 0b100:
			return Decoded{XOR, InstructionFormatR}
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				return Decoded{SRL, InstructionFormatR}
			case 0b010_0000:
				return Decoded{SRA, InstructionFormatR}
			}
		case 0b110:
			return Decoded{OR, InstructionFormatR}
		case 0b111:
			return Decoded{AND, InstructionFormatR}
		}
	case 0b110_0111:
		return Decoded{JALR, InstructionFormatI}
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			return Decoded{LB, InstructionFormatI}
		case 0b001:
			return Decoded{LH, InstructionFormatI}
		case 0b010:
			return Decoded{LW, InstructionFormatI}
		case 0b100:
			return Decoded{LBU, InstructionFormatI}
		case 0b101:
			return Decoded{LHU, InstructionFormatI}
		}
	case 0b001_0011:
		switch funct3 {
		case 0b000:
			return Decoded{ADDI, InstructionFormatI}
		case 0b010:
			return Decoded{SLTI, InstructionFormatI}
		case 0b011:
			return Decoded{SLTIU, InstructionFormatI}
		case 0b100:
			return Decoded{XORI, InstructionFormatI}
		case 0b110:
			return Decoded{ORI, InstructionFormatI}
		case 0b111:
			return Decoded{ANDI, InstructionFormatI}
		case 0b001:
			return Decoded{SLLI, InstructionFormatI}
		case 0b101:
			imm := bits(inst, 31, 20)
			switch imm >> 5 {
			case 0b000_0000:
				return Decoded{SRLI, InstructionFormatI}
			case 0b010_0000:
				return Decoded{SRAI, InstructionFormatI}
			}
		}
	case 0b000_1111:
		switch funct3 {
		case 0b000:
			return Decoded{FENCE, InstructionFormatI}
		case 0b001:
			return Decoded{FENCE_I, InstructionFormatI}
		}
	case 0b111_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_1000:
				switch bits(inst, 24, 20) {
				case 0b0_0010:
					return Decoded{SRET, InstructionFormatR}
				case 0b0_0101:
					return Decoded{WFI, InstructionFormatR}
				}
			case 0b001_1000:
				return Decoded{MRET, InstructionFormatR}
			case 0b000_1001:
				return Decoded{SFENCE_VMA, InstructionFormatR}
			case 0b000_0000:
				imm := bits(inst, 24, 20)
				switch imm {
				case 0b00:
					return Decoded{ECALL, InstructionFormatI}
				case 0b01:
					return Decoded{EBREAK, InstructionFormatI}
				case 0b10:
					return Decoded{URET, InstructionFormatR}
				}
			}
		case 0b001:
			return Decoded{CSRRW, InstructionFormatI}
		case 0b010:
			return Decoded{CSRRS, InstructionFormatI}
		case 0b011:
			return Decoded{CSRRC, InstructionFormatI}
		case 0b101:
			return Decoded{CSRRWI, InstructionFormatI}
		case 0b110:
			return Decoded{CSRRSI, InstructionFormatI}
		case 0b111:
			return Decoded{CSRRCI, InstructionFormatI}
		}
	case 0b001_0111:
		return Decoded{AUIPC, InstructionFormatU}
	case 0b011_0111:
		return Decoded{LUI, InstructionFormatU}
	case 0b110_1111:
		return Decoded{JAL, InstructionFormatJ}
	case 0b010_0011:
		switch funct3 {
		case 0b000:
			return Decoded{SB, InstructionFormatS}
		case 0b001:
			return Decoded{SH, InstructionFormatS}
		case 0b010:
			return Decoded{SW, InstructionFormatS}
		}
	case 0b110_0011:
		switch funct3 {
		case 0b000:
			return Decoded{BEQ, InstructionFormatB}
		case 0b001:
			return Decoded{BNE, InstructionFormatB}
		case 0b100:
			return Decoded{BLT, InstructionFormatB}
		case 0b101:
			return Decoded{BGE, InstructionFormatB}
		case 0b110:
			return Decoded{BLTU, InstructionFormatB}
		case 0b111:
			return Decoded{BGEU, InstructionFormatB}
		}
	}

	return Decoded{_INVALID, InstructionFormatInvalid}
}
