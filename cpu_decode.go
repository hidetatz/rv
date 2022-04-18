package main

type InstructionParam interface {
	isInstParam()
}

type Decoded struct {
	Code  InstructionCode
	Param InstructionParam
}

// Decode returns the instruction to be executed.
func (cpu *CPU) Decode(inst uint64) (*Decoded, Exception) {
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
				return &Decoded{ADD, ParseR(inst)}, ExcpNone
			case 0b010_0000:
				return &Decoded{SUB, ParseR(inst)}, ExcpNone
			}
		case 0b001:
			return &Decoded{SLL, ParseR(inst)}, ExcpNone
		case 0b010:
			return &Decoded{SLT, ParseR(inst)}, ExcpNone
		case 0b011:
			return &Decoded{SLTU, ParseR(inst)}, ExcpNone
		case 0b100:
			return &Decoded{XOR, ParseR(inst)}, ExcpNone
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				return &Decoded{SRL, ParseR(inst)}, ExcpNone
			case 0b010_0000:
				return &Decoded{SRA, ParseR(inst)}, ExcpNone
			}
		case 0b110:
			return &Decoded{OR, ParseR(inst)}, ExcpNone
		case 0b111:
			return &Decoded{AND, ParseR(inst)}, ExcpNone
		}
	case 0b110_0111:
		return &Decoded{JALR, ParseI(inst)}, ExcpNone
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			return &Decoded{LB, ParseI(inst)}, ExcpNone
		case 0b001:
			return &Decoded{LH, ParseI(inst)}, ExcpNone
		case 0b010:
			return &Decoded{LW, ParseI(inst)}, ExcpNone
		case 0b100:
			return &Decoded{LBU, ParseI(inst)}, ExcpNone
		case 0b101:
			return &Decoded{LHU, ParseI(inst)}, ExcpNone
		}
	case 0b001_0011:
		switch funct3 {
		case 0b000:
			return &Decoded{ADDI, ParseI(inst)}, ExcpNone
		case 0b010:
			return &Decoded{SLTI, ParseI(inst)}, ExcpNone
		case 0b011:
			return &Decoded{SLTIU, ParseI(inst)}, ExcpNone
		case 0b100:
			return &Decoded{XORI, ParseI(inst)}, ExcpNone
		case 0b110:
			return &Decoded{ORI, ParseI(inst)}, ExcpNone
		case 0b111:
			return &Decoded{ANDI, ParseI(inst)}, ExcpNone
		case 0b001:
			return &Decoded{SLLI, ParseI(inst)}, ExcpNone
		case 0b101:
			imm := bits(inst, 31, 20)
			switch imm >> 5 {
			case 0b000_0000:
				return &Decoded{SRLI, ParseI(inst)}, ExcpNone
			case 0b010_0000:
				return &Decoded{SRAI, ParseI(inst)}, ExcpNone
			}
		}
	case 0b000_1111:
		switch funct3 {
		case 0b000:
			return &Decoded{FENCE, ParseI(inst)}, ExcpNone
		case 0b001:
			return &Decoded{FENCE_I, ParseI(inst)}, ExcpNone
		}
	case 0b111_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_1000:
				switch bits(inst, 24, 20) {
				case 0b0_0010:
					return &Decoded{SRET, ParseR(inst)}, ExcpNone
				case 0b0_0101:
					return &Decoded{WFI, ParseR(inst)}, ExcpNone
				}
			case 0b001_1000:
				return &Decoded{MRET, ParseR(inst)}, ExcpNone
			case 0b000_1001:
				return &Decoded{SFENCE_VMA, ParseR(inst)}, ExcpNone
			case 0b000_0000:
				imm := bits(inst, 24, 20)
				switch imm {
				case 0b00:
					return &Decoded{ECALL, ParseI(inst)}, ExcpNone
				case 0b01:
					return &Decoded{EBREAK, ParseI(inst)}, ExcpNone
				case 0b10:
					return &Decoded{URET, ParseR(inst)}, ExcpNone
				}
			}
		case 0b001:
			return &Decoded{CSRRW, ParseI(inst)}, ExcpNone
		case 0b010:
			return &Decoded{CSRRS, ParseI(inst)}, ExcpNone
		case 0b011:
			return &Decoded{CSRRC, ParseI(inst)}, ExcpNone
		case 0b101:
			return &Decoded{CSRRWI, ParseI(inst)}, ExcpNone
		case 0b110:
			return &Decoded{CSRRSI, ParseI(inst)}, ExcpNone
		case 0b111:
			return &Decoded{CSRRCI, ParseI(inst)}, ExcpNone
		}
	case 0b001_0111:
		return &Decoded{AUIPC, ParseU(inst)}, ExcpNone
	case 0b011_0111:
		return &Decoded{LUI, ParseU(inst)}, ExcpNone
	case 0b110_1111:
		return &Decoded{JAL, ParseJ(inst)}, ExcpNone
	case 0b010_0011:
		switch funct3 {
		case 0b000:
			return &Decoded{SB, ParseS(inst)}, ExcpNone
		case 0b001:
			return &Decoded{SH, ParseS(inst)}, ExcpNone
		case 0b010:
			return &Decoded{SW, ParseS(inst)}, ExcpNone
		}
	case 0b110_0011:
		switch funct3 {
		case 0b000:
			return &Decoded{BEQ, ParseB(inst)}, ExcpNone
		case 0b001:
			return &Decoded{BNE, ParseB(inst)}, ExcpNone
		case 0b100:
			return &Decoded{BLT, ParseB(inst)}, ExcpNone
		case 0b101:
			return &Decoded{BGE, ParseB(inst)}, ExcpNone
		case 0b110:
			return &Decoded{BLTU, ParseB(inst)}, ExcpNone
		case 0b111:
			return &Decoded{BGEU, ParseB(inst)}, ExcpNone
		}
	}

	return nil, ExcpIllegalInstruction
}
