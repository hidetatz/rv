package main

import "math"

type InstructionCode string

const (
	_INVALID = InstructionCode("_INVALID")

	// RV32C
	C_ADDI4SPN = InstructionCode("C.ADDI4SPN")
	C_FLD      = InstructionCode("C.FLD")
	C_LW       = InstructionCode("C.LW")
	C_FLW      = InstructionCode("C.FLW")
	C_FSD      = InstructionCode("C.FSD")
	C_SW       = InstructionCode("C.SW")
	C_FSW      = InstructionCode("C.FSW")
	C_ADDI16SP = InstructionCode("C.ADDI16SP")
	C_LUI      = InstructionCode("C.LUI")
	C_JAL      = InstructionCode("C.JAL")
	C_LI       = InstructionCode("C.LI")
	C_NOP      = InstructionCode("C.NOP")
	C_ADDI     = InstructionCode("C.ADDI")
	C_SRLI     = InstructionCode("C.SRLI")
	C_SRAI     = InstructionCode("C.SRAI")
	C_ANDI     = InstructionCode("C.ANDI")
	C_SUB      = InstructionCode("C.SUB")
	C_XOR      = InstructionCode("C.XOR")
	C_OR       = InstructionCode("C.OR")
	C_AND      = InstructionCode("C.AND")
	C_J        = InstructionCode("C.J")
	C_BEQZ     = InstructionCode("C.BEQZ")
	C_BNEZ     = InstructionCode("C.BNEZ")
	C_SLLI     = InstructionCode("C.SLLI")
	C_FLDSP    = InstructionCode("C.FLDSP")
	C_LWSP     = InstructionCode("C.LWSP")
	C_FLWSP    = InstructionCode("C.FLWSP")
	C_JR       = InstructionCode("C.JR")
	C_MV       = InstructionCode("C.MV")
	C_EBREAK   = InstructionCode("C.EBREAK")
	C_JALR     = InstructionCode("C.JALR")
	C_ADD      = InstructionCode("C.ADD")
	C_FSDSP    = InstructionCode("C.FSDSP")
	C_SWSP     = InstructionCode("C.SWSP")
	C_FSWSP    = InstructionCode("C.FSWSP")

	// RV32I, RV64I
	LUI        = InstructionCode("LUI")
	AUIPC      = InstructionCode("AUIPC")
	ADDI       = InstructionCode("ADDI")
	SLTI       = InstructionCode("SLTI")
	SLTIU      = InstructionCode("SLTIU")
	XORI       = InstructionCode("XORI")
	ORI        = InstructionCode("ORI")
	ANDI       = InstructionCode("ANDI")
	SLLI       = InstructionCode("SLLI")
	SRLI       = InstructionCode("SRLI")
	SRAI       = InstructionCode("SRAI")
	ADD        = InstructionCode("ADD")
	SUB        = InstructionCode("SUB")
	SLL        = InstructionCode("SLL")
	SLT        = InstructionCode("SLT")
	SLTU       = InstructionCode("SLTU")
	XOR        = InstructionCode("XOR")
	SRL        = InstructionCode("SRL")
	SRA        = InstructionCode("SRA")
	OR         = InstructionCode("OR")
	AND        = InstructionCode("AND")
	FENCE      = InstructionCode("FENCE")
	FENCE_I    = InstructionCode("FENCE.I")
	CSRRW      = InstructionCode("CSRRW")
	CSRRS      = InstructionCode("CSRRS")
	CSRRC      = InstructionCode("CSRRC")
	CSRRWI     = InstructionCode("CSRRWI")
	CSRRSI     = InstructionCode("CSRRSI")
	CSRRCI     = InstructionCode("CSRRCI")
	ECALL      = InstructionCode("ECALL")
	EBREAK     = InstructionCode("EBREAK")
	URET       = InstructionCode("URET")
	SRET       = InstructionCode("SRET")
	MRET       = InstructionCode("MRET")
	WFI        = InstructionCode("WFI")
	SFENCE_VMA = InstructionCode("SFENCE.VMA")
	LB         = InstructionCode("LB")
	LH         = InstructionCode("LH")
	LW         = InstructionCode("LW")
	LBU        = InstructionCode("LBU")
	LHU        = InstructionCode("LHU")
	SB         = InstructionCode("SB")
	SH         = InstructionCode("SH")
	SW         = InstructionCode("SW")
	JAL        = InstructionCode("JAL")
	JALR       = InstructionCode("JALR")
	BEQ        = InstructionCode("BEQ")
	BNE        = InstructionCode("BNE")
	BLT        = InstructionCode("BLT")
	BGE        = InstructionCode("BGE")
	BLTU       = InstructionCode("BLTU")
	BGEU       = InstructionCode("BGEU")

	// RV64I
	//ADDIW = InstructionCode("ADDIW")
	//SLLIW = InstructionCode("SLLIW")
	//SRLIW = InstructionCode("SRLIW")
	//SRAIW = InstructionCode("SRAIW")
	//ADDW  = InstructionCode("ADDW")
	//SUBW  = InstructionCode("SUBW")
	//SLLW = InstructionCode("SLLW")
	//SRLW  = InstructionCode("SRLW")
	//SRAW  = InstructionCode("SRAW")
	//LWU = InstructionCode("LWU")
	//LD    = InstructionCode("LD")
	//SD    = InstructionCode("SD")
)

func (ic InstructionCode) String() string {
	return string(ic)
}

// Instructions is the mapping of InstructionCode and the operation to be executed.
// raw is the instruction machine code, which will be either 32-bit or 16-bit (if compressed).
// pc is the program counter at which the instruction is fetched.
// Note that cpu.PC is not the point the instruction is fetched, because it is already incremented on for the next iteration.
var Instructions = map[InstructionCode]func(cpu *CPU, raw, pc uint64) Exception{
	/*
	 * RV32C
	 */
	C_ADDI4SPN: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 4, 2) + 8
		nzuimm := (bits(raw, 12, 11) << 4) | // raw[12:11] -> nzuimm[5:4]
			(bits(raw, 10, 7) << 6) | // raw[10:7] -> nzuimm[9:6]
			(bit(raw, 6) << 2) | // raw[6] -> nzuimm[2]
			(bit(raw, 5) << 3) // raw[5] -> nzuimm[3]

		if nzuimm == 0 {
			return ExcpIllegalInstruction
		}

		cpu.XRegs.Write(rd, cpu.XRegs.Read(2)+uint64(nzuimm))
		return ExcpNone
	},
	C_FLD: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bits(raw, 6, 5) << 6) // raw[6:5] -> offset[7:6]

		v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+uint64(offset), DoubleWord)
		cpu.FRegs.Write(rd, math.Float64frombits(v))
		return ExcpNone
	},
	C_LW: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]

		addr := cpu.XRegs.Read(rs1) + offset
		v := cpu.Bus.Read(addr, Word)
		cpu.XRegs.Write(rd, uint64(int64(int32(v))))
		return ExcpNone
	},
	C_FLW: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, Word)
		cpu.FRegs.Write(rd, math.Float64frombits(v))
		return ExcpNone
	},
	C_FSD: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 7) | // raw[6] -> offset[7]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		addr := cpu.XRegs.Read(rs1) + offset
		cpu.Bus.Write(addr, math.Float64bits(cpu.FRegs.Read(rd)), DoubleWord)
		return ExcpNone
	},
	C_SW: func(cpu *CPU, raw, _ uint64) Exception {
		rs1 := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		addr := cpu.XRegs.Read(rs1) + offset
		cpu.Bus.Write(addr, cpu.XRegs.Read(rs2), Word)
		return ExcpNone
	},
	C_FSW: func(cpu *CPU, raw, _ uint64) Exception {
		rs1 := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		addr := cpu.XRegs.Read(rs1) + offset
		cpu.Bus.Write(addr, math.Float64bits(cpu.FRegs.Read(rs2)), Word)
		return ExcpNone
	},
	C_ADDI16SP: func(cpu *CPU, raw, _ uint64) Exception {
		imm := (bit(raw, 12) << 9) | // raw[12] -> imm[9]
			(bit(raw, 6) << 4) | // raw[6] -> imm[4]
			(bit(raw, 5) << 6) | // raw[5] -> imm[6]
			(bits(raw, 4, 3) << 7) | // raw[4:3] -> imm[8:7]
			(bit(raw, 2) << 5) // raw[2] -> imm[5]
		if (imm & 0b10_0000_0000) != 0 {
			// sign-extend
			imm = uint64(int64(int32(int16((imm | 0b1111_1100_0000_0000)))))
		}

		if imm == 0 {
			return ExcpNone
		}

		// write to stack pointer (x2)
		cpu.XRegs.Write(2, cpu.XRegs.Read(2)+imm)
		return ExcpNone
	},
	C_LUI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 11, 7)
		if rd == 2 {
			return ExcpNone
		}

		imm := (bit(raw, 12) << 17) | // raw[12] -> imm[17]
			(bits(raw, 6, 2) << 12) // raw[6:2] -> imm[16:12]
		if (imm & 0b10_0000_0000_0000_0000) != 0 {
			// sign-extend
			imm = uint64(int64(int32((imm | 0b1111_1111_1111_1100_0000_0000_0000_0000))))
		}
		if imm == 0 {
			return ExcpNone
		}

		// write to stack pointer (x2)
		cpu.XRegs.Write(rd, imm)
		return ExcpNone
	},
	C_JAL: func(cpu *CPU, raw, pc uint64) Exception {
		offset := (bit(raw, 12) << 11) | // raw[12] -> imm[11]
			(bit(raw, 11) << 4) | // raw[11] -> imm[4]
			(bit(raw, 10) << 9) | // raw[11] -> imm[4]
			(bit(raw, 9) << 8) | // raw[11] -> imm[4]
			(bit(raw, 8) << 10) | // raw[11] -> imm[4]
			(bit(raw, 7) << 6) | // raw[11] -> imm[4]
			(bit(raw, 6) << 7) | // raw[11] -> imm[4]
			(bit(raw, 5) << 3) | // raw[11] -> imm[4]
			(bit(raw, 4) << 2) | // raw[11] -> imm[4]
			(bit(raw, 3) << 1) | // raw[11] -> imm[4]
			(bit(raw, 2) << 5) // raw[11] -> imm[4]

		if (offset & 0b1_0000_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1110_0000_0000_0000))))
		}

		tmp := pc + 2           // use 2 because compressed
		cpu.XRegs.Write(1, tmp) // use x1
		cpu.PC = pc + offset
		return ExcpNone
	},
	C_LI: func(cpu *CPU, raw, pc uint64) Exception {
		rd := bits(raw, 11, 7)
		imm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (imm & 0b10_0000) != 0 {
			// sign-extend
			imm = uint64(int64(int32(int16(imm | 0b1111_1111_1100_0000))))
		}

		cpu.XRegs.Write(rd, (imm + cpu.XRegs.Read(0)))
		return ExcpNone
	},
	C_NOP: func(cpu *CPU, raw, _ uint64) Exception {
		// nop does nothing
		return ExcpNone
	},
	C_ADDI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 11, 7)
		nzimm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (nzimm & 0b10_0000) != 0 {
			// sign-extend
			nzimm = uint64(int64(int32(int16(nzimm | 0b1111_1111_1100_0000))))
		}
		cpu.XRegs.Write(rd, (nzimm + cpu.XRegs.Read(rd)))
		return ExcpNone
	},
	C_SRLI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) >> shamt))
		return ExcpNone
	},
	C_SRAI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.XRegs.Write(rd, uint64(int64(cpu.XRegs.Read(rd))>>shamt))
		return ExcpNone
	},
	C_ANDI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		uimm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (uimm & 0b10_0000) != 0 {
			// sign-extend
			uimm = uint64(int64(int32(int16(uimm | 0b1111_1111_1100_0000))))
		}
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) & uimm))
		return ExcpNone
	},
	C_SUB: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) - cpu.XRegs.Read(rs2)))
		return ExcpNone
	},
	C_XOR: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) ^ cpu.XRegs.Read(rs2)))
		return ExcpNone
	},
	C_OR: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) | cpu.XRegs.Read(rs2)))
		return ExcpNone
	},
	C_AND: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) & cpu.XRegs.Read(rs2)))
		return ExcpNone
	},
	C_J: func(cpu *CPU, raw, pc uint64) Exception {
		offset := (bit(raw, 12) << 11) | // raw[12] -> imm[11]
			(bit(raw, 11) << 4) | // raw[11] -> imm[4]
			(bit(raw, 10) << 9) | // raw[11] -> imm[4]
			(bit(raw, 9) << 8) | // raw[11] -> imm[4]
			(bit(raw, 8) << 10) | // raw[11] -> imm[4]
			(bit(raw, 7) << 6) | // raw[11] -> imm[4]
			(bit(raw, 6) << 7) | // raw[11] -> imm[4]
			(bit(raw, 5) << 3) | // raw[11] -> imm[4]
			(bit(raw, 4) << 2) | // raw[11] -> imm[4]
			(bit(raw, 3) << 1) | // raw[11] -> imm[4]
			(bit(raw, 2) << 5) // raw[11] -> imm[4]
		if (offset & 0b1_0000_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1110_0000_0000_0000))))
		}
		cpu.PC = pc + offset
		return ExcpNone
	},
	C_BEQZ: func(cpu *CPU, raw, pc uint64) Exception {
		rs1 := bits(raw, 9, 7) + 8
		offset := (bit(raw, 12) << 8) |
			(bits(raw, 11, 10) << 3) |
			(bits(raw, 6, 5) << 6) |
			(bits(raw, 4, 3) << 1) |
			(bit(raw, 2) << 5)
		if (offset & 0b1_0000_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1110_0000_0000_0000))))
		}

		if cpu.XRegs.Read(rs1) == 0 {
			cpu.PC = pc + offset
		}

		return ExcpNone
	},
	C_BNEZ: func(cpu *CPU, raw, pc uint64) Exception {
		rs1 := bits(raw, 9, 7) + 8
		offset := (bit(raw, 12) << 8) |
			(bits(raw, 11, 10) << 3) |
			(bits(raw, 6, 5) << 6) |
			(bits(raw, 4, 3) << 1) |
			(bit(raw, 2) << 5)
		if (offset & 0b1_0000_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1110_0000_0000_0000))))
		}

		if cpu.XRegs.Read(rs1) != 0 {
			cpu.PC = pc + offset
		}

		return ExcpNone
	},
	C_SLLI: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 11, 7)
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.XRegs.Write(rd, (cpu.XRegs.Read(rd) << shamt))
		return ExcpNone
	},
	C_FLDSP: func(cpu *CPU, raw, _ uint64) Exception {
		rd := bits(raw, 11, 7)
		uimm := (bit(raw, 12) << 5) | (bits(raw, 6, 5) << 3) | (bits(raw, 4, 2) << 6)
		v := math.Float64frombits(cpu.Bus.Read(cpu.XRegs.Read(2)+uimm, DoubleWord))
		cpu.FRegs.Write(rd, v)
		return ExcpNone
	},

	//C_XX: func(cpu *CPU, raw, _ uint64) Exception {
	//	return ExcpNone
	//},

	/*
	 * RV64I
	 */
	ADD: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)+cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SUB: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)-cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SLL: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		shamt := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)<<shamt)
		return ExcpNone
	},
	SLT: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		var v uint64 = 0
		if int64(cpu.XRegs.Read(i.Rs1)) < int64(cpu.XRegs.Read(i.Rs2)) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	SLTU: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		var v uint64 = 0
		if cpu.XRegs.Read(i.Rs1) < cpu.XRegs.Read(i.Rs2) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	XOR: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)^cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SRL: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		shift := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)>>shift)
		return ExcpNone
	},
	SRA: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		shift := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, uint64(int64(cpu.XRegs.Read(i.Rs1))>>shift))
		return ExcpNone
	},
	OR: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)|cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	AND: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)&cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SFENCE_VMA: func(cpu *CPU, raw, _ uint64) Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone
	},
	WFI: func(cpu *CPU, raw, _ uint64) Exception {
		cpu.Wfi = true
		return ExcpNone
	},
	SRET: func(cpu *CPU, raw, _ uint64) Exception {
		// First, set CSRs[SEPC] to program counter.
		cpu.PC = cpu.CSR.Read(CsrSEPC)

		// Then, Modify SSTATUS.

		sstatus := cpu.CSR.Read(CsrSSTATUS)

		// Set CPU mode according to SPP
		switch bit(sstatus, CsrStatusSPP) {
		case 0b0:
			cpu.Mode = User
		case 0b1:
			cpu.Mode = Supervisor
		default:
			// should not happen
			panic("invalid CSR SPP")
		}

		spie := bit(sstatus, CsrStatusSPIE)

		// set SPIE to SIE
		if spie == 0 {
			sstatus = clearBit(sstatus, CsrStatusSIE)
		} else {
			sstatus = setBit(sstatus, CsrStatusSIE)
		}

		// set 1 to SPIE
		sstatus = setBit(sstatus, CsrStatusSPIE)

		// set 0 to SPP
		sstatus = clearBit(sstatus, CsrStatusSPP)

		// update SSTATUS
		cpu.CSR.Write(CsrSSTATUS, sstatus)

		// MPRV must be set 0 if the mode is not Machine.
		if cpu.Mode == Supervisor {
			mstatus := cpu.CSR.Read(CsrMSTATUS)
			mstatus = clearBit(mstatus, CsrStatusMPRV)
			cpu.CSR.Write(CsrMSTATUS, mstatus)
		}

		return ExcpNone
	},
	MRET: func(cpu *CPU, raw, _ uint64) Exception {
		// First, set CSRs[MEPC] to program counter.
		cpu.PC = cpu.CSR.Read(CsrMEPC)

		// Then, Modify MSTATUS.

		mstatus := cpu.CSR.Read(CsrMSTATUS)

		// Set CPU mode according to MPP
		switch bits(mstatus, CsrStatusMPPHi, CsrStatusMPPLo) {
		case 0b00:
			cpu.Mode = User
			mstatus = clearBit(mstatus, CsrStatusMPRV)
		case 0b01:
			cpu.Mode = Supervisor
			mstatus = clearBit(mstatus, CsrStatusMPRV)
		case 0b11:
			cpu.Mode = Machine
		default:
			// should not happen
			panic("invalid CSR MPP")
		}

		mpie := bit(mstatus, CsrStatusMPIE)

		// set MPIE to MIE
		if mpie == 0 {
			mstatus = clearBit(mstatus, CsrStatusMIE)
		} else {
			mstatus = setBit(mstatus, CsrStatusMIE)
		}

		// set 1 to MPIE
		mstatus = setBit(mstatus, CsrStatusMPIE)

		// set 0 to MPP
		mstatus = clearBit(mstatus, CsrStatusMPPHi)
		mstatus = clearBit(mstatus, CsrStatusMPPLo)

		// update MSTATUS
		cpu.CSR.Write(CsrMSTATUS, mstatus)

		return ExcpNone
	},
	URET: func(cpu *CPU, raw, _ uint64) Exception {
		ustatus := cpu.CSR.Read(CsrUSTATUS)
		upie := bit(ustatus, CsrStatusUPIE)

		// set UPIE to UIE
		if upie == 0 {
			ustatus = clearBit(ustatus, CsrStatusUIE)
		} else {
			ustatus = setBit(ustatus, CsrStatusUIE)
		}

		// set 1 to SPIE
		ustatus = setBit(ustatus, CsrStatusUPIE)

		// update USTATUS
		cpu.CSR.Write(CsrUSTATUS, ustatus)

		return ExcpNone
	},
	JALR: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseI(raw)
		tmp := pc + 4
		target := (cpu.XRegs.Read(i.Rs1) + i.Imm) & ^uint64(1)
		cpu.PC = target - 4 // sub in advance as the PC is incremented later
		cpu.XRegs.Write(i.Rd, tmp)
		return ExcpNone
	},
	LB: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 8)
		cpu.XRegs.Write(i.Rd, uint64(int64(int8(v))))
		return ExcpNone
	},
	LH: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 16)
		cpu.XRegs.Write(i.Rd, uint64(int64(int16(v))))
		return ExcpNone
	},
	LW: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 32)
		cpu.XRegs.Write(i.Rd, uint64(int64(int32(v))))
		return ExcpNone
	},
	LBU: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 8)
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	LHU: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 16)
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	ADDI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, i.Imm+cpu.XRegs.Read(i.Rs1))
		return ExcpNone
	},
	SLTI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.XRegs.Read(i.Rs1)) < int64(i.Imm) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	SLTIU: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		var v uint64 = 0
		// must compare as two's complement
		if cpu.XRegs.Read(i.Rs1) < i.Imm {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	XORI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)^i.Imm)
		return ExcpNone
	},
	ORI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)|i.Imm)
		return ExcpNone
	},
	ANDI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)&i.Imm)
		return ExcpNone
	},
	SLLI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)<<shamt)
		return ExcpNone
	},
	SRLI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)>>shamt)
		return ExcpNone
	},
	SRAI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, uint64(int64(cpu.XRegs.Read(i.Rs1))>>shamt))
		return ExcpNone
	},
	FENCE: func(cpu *CPU, raw, _ uint64) Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone
	},
	FENCE_I: func(cpu *CPU, raw, _ uint64) Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone
	},
	ECALL: func(cpu *CPU, raw, _ uint64) Exception {
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
		return ExcpIllegalInstruction
	},
	EBREAK: func(cpu *CPU, raw, _ uint64) Exception {
		return ExcpBreakpoint
	},
	CSRRW: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		t := cpu.CSR.Read(i.Imm)
		cpu.CSR.Write(i.Imm, cpu.XRegs.Read(i.Rs1))
		cpu.XRegs.Write(i.Rd, t)

		return ExcpNone
	},
	CSRRS: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		t := cpu.CSR.Read(i.Imm)
		cpu.CSR.Write(i.Imm, (t | cpu.XRegs.Read(i.Rs1)))
		cpu.XRegs.Write(i.Rd, t)

		return ExcpNone
	},
	CSRRC: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		t := cpu.CSR.Read(i.Imm)
		cpu.CSR.Write(i.Imm, (t & ^(cpu.XRegs.Read(i.Rs1))))
		cpu.XRegs.Write(i.Rd, t)

		return ExcpNone
	},
	CSRRWI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.CSR.Read(i.Imm))
		cpu.CSR.Write(i.Imm, i.Rs1) // RS1 is zimm

		return ExcpNone
	},
	CSRRSI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		t := cpu.CSR.Read(i.Imm)
		cpu.CSR.Write(i.Imm, (t | i.Rs1)) // RS1 is zimm
		cpu.XRegs.Write(i.Rd, t)

		return ExcpNone
	},
	CSRRCI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseI(raw)
		t := cpu.CSR.Read(i.Imm)
		cpu.CSR.Write(i.Imm, (t & ^(i.Rs1)))
		cpu.XRegs.Write(i.Rd, t)

		return ExcpNone
	},
	AUIPC: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseU(raw)
		cpu.XRegs.Write(i.Rd, pc+i.Imm)
		return ExcpNone
	},
	LUI: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseU(raw)
		cpu.XRegs.Write(i.Rd, i.Imm)
		return ExcpNone
	},
	SB: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseS(raw)
		pc := cpu.XRegs.Read(i.Rs1) + i.Imm
		cpu.Bus.Write(pc, cpu.XRegs.Read(i.Rs2), Byte)
		return ExcpNone
	},
	SH: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseS(raw)
		pc := cpu.XRegs.Read(i.Rs1) + i.Imm
		cpu.Bus.Write(pc, cpu.XRegs.Read(i.Rs2), HalfWord)
		return ExcpNone
	},
	SW: func(cpu *CPU, raw, _ uint64) Exception {
		i := ParseS(raw)
		pc := cpu.XRegs.Read(i.Rs1) + i.Imm
		cpu.Bus.Write(pc, cpu.XRegs.Read(i.Rs2), Word)
		return ExcpNone
	},
	BEQ: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if cpu.XRegs.Read(i.Rs1) == cpu.XRegs.Read(i.Rs2) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	BNE: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if cpu.XRegs.Read(i.Rs1) != cpu.XRegs.Read(i.Rs2) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	BLT: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if int64(cpu.XRegs.Read(i.Rs1)) < int64(cpu.XRegs.Read(i.Rs2)) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	BGE: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if int64(cpu.XRegs.Read(i.Rs1)) >= int64(cpu.XRegs.Read(i.Rs2)) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	BLTU: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if cpu.XRegs.Read(i.Rs1) < cpu.XRegs.Read(i.Rs2) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	BGEU: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseB(raw)
		if cpu.XRegs.Read(i.Rs1) >= cpu.XRegs.Read(i.Rs2) {
			cpu.PC = pc + i.Imm
		}
		return ExcpNone
	},
	JAL: func(cpu *CPU, raw, pc uint64) Exception {
		i := ParseJ(raw)
		tmp := pc + 4
		if i.Rd == 0b0 {
			i.Rd = 1 // x1 if rd is omitted
		}
		cpu.XRegs.Write(i.Rd, tmp)
		cpu.PC = pc + i.Imm
		return ExcpNone
	},
}
