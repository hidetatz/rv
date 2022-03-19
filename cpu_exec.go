package main

// bits returns val[hi:lo].
func bits(val uint64, hi, lo int) uint64 {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

func (cpu *CPU) Exec(format InstructionFormat, inst uint64) Exception {
	switch format {
	case InstructionFormatR:
		return cpu.ExecR(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 14, 12),
			bits(inst, 19, 15),
			bits(inst, 24, 20),
			bits(inst, 31, 25),
		)
	case InstructionFormatI:
		return cpu.ExecI(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 14, 12),
			bits(inst, 19, 15),
			bits(inst, 31, 20),
		)
	case InstructionFormatS:
		return cpu.ExecS(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 14, 12),
			bits(inst, 19, 15),
			bits(inst, 24, 20),
			bits(inst, 31, 25),
		)
	case InstructionFormatB:
		return cpu.ExecB(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 14, 12),
			bits(inst, 19, 15),
			bits(inst, 24, 20),
			bits(inst, 31, 25),
		)
	case InstructionFormatU:
		return cpu.ExecU(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 31, 12),
		)
	case InstructionFormatJ:
		return cpu.ExecJ(
			bits(inst, 6, 0),
			bits(inst, 11, 7),
			bits(inst, 31, 12),
		)
	default:
		return ExcpIllegalInstruction
	}
}

func (cpu *CPU) ExecR(opcode, rd, funct3, rs1, rs2, funct7 uint64) Exception {
	switch opcode {
	case 0b011_0011: // RV32I
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				// add
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)+cpu.XRegs.Read(rs2))
			case 0b010_0000:
				// sub
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)-cpu.XRegs.Read(rs2))
			default:
				return ExcpIllegalInstruction
			}
		case 0b001:
			// sll
			// In RV64I, only the low 6 bits of rs2 are used as the shift amount
			shamt := cpu.XRegs.Read(rs2) & 0b111111
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)<<shamt)
		case 0b010:
			// slt
			var v uint64 = 0
			if int64(cpu.XRegs.Read(rs1)) < int64(cpu.XRegs.Read(rs2)) {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b011:
			// sltu
			var v uint64 = 0
			if cpu.XRegs.Read(rs1) < cpu.XRegs.Read(rs2) {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b100:
			// xor
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)^cpu.XRegs.Read(rs2))
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				// srl
				shift := cpu.XRegs.Read(rs2) & 0b111111
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)>>shift)
			case 0b010_0000:
				// sra
				shift := cpu.XRegs.Read(rs2) & 0b111111
				cpu.XRegs.Write(rd, uint64(int64(cpu.XRegs.Read(rs1))>>shift))
			default:
				return ExcpIllegalInstruction
			}
		case 0b110:
			// or
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)|cpu.XRegs.Read(rs2))
		case 0b111:
			// and
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)&cpu.XRegs.Read(rs2))
		default:
			return ExcpIllegalInstruction
		}
	default:
		return ExcpIllegalInstruction
	}

	return ExcpNone
}

func (cpu *CPU) ExecI(opcode, rd, funct3, rs1, imm uint64) Exception {
	switch opcode {
	case 0b110_0111: // RV32I
		// jalr
		tmp := cpu.PC + 4
		target := (cpu.XRegs.Read(rs1) + imm) & ^uint64(1)
		cpu.PC = target - 4 // sub in advance as the PC is incremented later
		cpu.XRegs.Write(rd, tmp)
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			// lb
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+imm, 8)
			cpu.XRegs.Write(rd, uint64(int64(int8(v))))
		case 0b001:
			// lh
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+imm, 16)
			cpu.XRegs.Write(rd, uint64(int64(int16(v))))
		case 0b010:
			// lw
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+imm, 32)
			cpu.XRegs.Write(rd, uint64(int64(int32(v))))
		case 0b100:
			// lbu
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+imm, 8)
			cpu.XRegs.Write(rd, v)
		case 0b101:
			// lhu
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+imm, 16)
			cpu.XRegs.Write(rd, v)
		default:
			return ExcpIllegalInstruction
		}
	case 0b001_0011:
		switch funct3 {
		case 0b000:
			// addi
			cpu.XRegs.Write(rd, imm+cpu.XRegs.Read(rs1))
		case 0b010:
			// slti
			var v uint64 = 0
			// must compare as two's complement
			if int64(cpu.XRegs.Read(rs1)) < int64(imm) {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b011:
			// sltiu
			var v uint64 = 0
			// must compare as two's complement
			if cpu.XRegs.Read(rs1) < imm {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b100:
			// xori
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)^imm)
		case 0b110:
			// ori
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)|imm)
		case 0b111:
			// andi
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)&imm)
		case 0b001:
			// slli
			shamt := imm & 0b1_1111
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)<<shamt)
		case 0b101:
			switch imm >> 5 {
			case 0b000_0000:
				// srli
				shamt := imm & 0b1_1111
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)>>shamt)
			case 0b010_0000:
				// srai
				shamt := imm & 0b1_1111
				cpu.XRegs.Write(rd, uint64(int64(cpu.XRegs.Read(rs1))>>shamt))
			default:
				return ExcpIllegalInstruction
			}
		default:
			return ExcpIllegalInstruction
		}
	case 0b000_1111: // RV32I
		switch funct3 {
		case 0b000:
			// fence
			// Do nothing because rv currently does not reorder the instructions for optimizations.
		case 0b001:
			// fence.i
			// Do nothing because rv currently does not reorder the instructions for optimizations.
		default:
			return ExcpIllegalInstruction
		}
	case 0b111_0011: // RV32I
		switch funct3 {
		case 0b000:
			switch imm {
			case 0b0:
				// ecall
				switch cpu.Mode {
				case User:
					return ExcpEnvironmentCallFromUmode
				case Supervisor:
					return ExcpEnvironmentCallFromSmode
				case Machine:
					return ExcpEnvironmentCallFromMmode
				default:
					return ExcpIllegalInstruction
				}
			case 0b1:
				// ebreak
				return ExcpBreakpoint
			default:
				return ExcpIllegalInstruction
			}
		case 0b001:
			// csrrw
		case 0b010:
			// csrrs
		case 0b011:
			// csrrc
		case 0b101:
			// csrr2wi
		case 0b110:
			// csrrsi
		case 0b111:
			// csrrci
		default:
			return ExcpIllegalInstruction
		}
	}
	return ExcpNone
}

func (cpu *CPU) ExecS(opcode, imm1, funct3, rs1, rs2, imm2 uint64) Exception {
	return ExcpNone
}

func (cpu *CPU) ExecB(opcode, imm1, funct3, rs1, rs2, imm2 uint64) Exception {
	return ExcpNone
}

func (cpu *CPU) ExecU(opcode, rd, imm uint64) Exception {
	switch opcode {
	case 0b001_0111: // RV32I
		// auipc
		cpu.XRegs.Write(rd, cpu.PC+imm)
	case 0b011_0111: // RV32I
		// lui
		cpu.XRegs.Write(rd, imm)
	}
	return ExcpNone
}

func (cpu *CPU) ExecJ(opcode, rd, imm uint64) Exception {
	switch opcode {
	case 0b110_1111: // RV32I
		// jal
		tmp := cpu.PC + 4
		if rd == 0b0 {
			rd = 1 // x1 if rd is omitted
		}
		cpu.XRegs.Write(rd, tmp)
		cpu.PC = imm - 4 // sub in advance as the PC is incremented later
	}
	return ExcpNone
}
