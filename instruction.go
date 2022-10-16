package main

import (
	"math"
	"math/big"
)

// InstructionCode is a psuedo-code which represents an instruction.
// The code itself internally doesn't have any meanings, it is just used to identify
// the instruction.
type InstructionCode string

const (
	_INVALID = InstructionCode("_INVALID")

	/*
	 * RV32C
	 */

	// Load
	C_LW    = InstructionCode("C.LW")
	C_LWSP  = InstructionCode("C.LWSP")
	C_FLD   = InstructionCode("C.FLD")
	C_FLDSP = InstructionCode("C.FLDSP")
	// Store
	C_SW    = InstructionCode("C.SW")
	C_SWSP  = InstructionCode("C.SWSP")
	C_FSD   = InstructionCode("C.FSD")
	C_FSDSP = InstructionCode("C.FSDSP")
	// Arithmetic
	C_ADD      = InstructionCode("C.ADD")
	C_ADDI     = InstructionCode("C.ADDI")
	C_ADDI4SPN = InstructionCode("C.ADDI4SPN")
	C_ADDI16SP = InstructionCode("C.ADDI16SP")
	C_SUB      = InstructionCode("C.SUB")
	C_AND      = InstructionCode("C.AND")
	C_ANDI     = InstructionCode("C.ANDI")
	C_OR       = InstructionCode("C.OR")
	C_XOR      = InstructionCode("C.XOR")
	C_MV       = InstructionCode("C.MV")
	C_LI       = InstructionCode("C.LI")
	C_LUI      = InstructionCode("C.LUI")
	// Shift
	C_SLLI = InstructionCode("C.SLLI")
	C_SRAI = InstructionCode("C.SRAI")
	C_SRLI = InstructionCode("C.SRLI")
	// Branch
	C_BEQZ = InstructionCode("C.BEQZ")
	C_BNEZ = InstructionCode("C.BNEZ")
	// Jump
	C_J  = InstructionCode("C.J")
	C_JR = InstructionCode("C.JR")
	// Jump and Link
	C_JALR = InstructionCode("C.JALR")
	// System
	C_NOP    = InstructionCode("C.NOP")
	C_EBREAK = InstructionCode("C.EBREAK")

	/*
	 * RV64C
	 */

	// Load
	C_LD   = InstructionCode("C.LD")
	C_LDSP = InstructionCode("C.LDSP")
	// Store
	C_SD   = InstructionCode("C.SD")
	C_SDSP = InstructionCode("C.SDSP")
	// Arithmetic
	C_ADDW  = InstructionCode("C.ADDW")
	C_ADDIW = InstructionCode("C.ADDIW")
	C_SUBW  = InstructionCode("C.SUBW")

	/*
	 * RV32I
	 */

	// Shift
	SLL  = InstructionCode("SLL")
	SLLI = InstructionCode("SLLI")
	SRL  = InstructionCode("SRL")
	SRLI = InstructionCode("SRLI")
	SRA  = InstructionCode("SRA")
	SRAI = InstructionCode("SRAI")
	// Arithmetic
	ADD   = InstructionCode("ADD")
	ADDI  = InstructionCode("ADDI")
	SUB   = InstructionCode("SUB")
	LUI   = InstructionCode("LUI")
	AUIPC = InstructionCode("AUIPC")
	// Logical
	XOR  = InstructionCode("XOR")
	XORI = InstructionCode("XORI")
	OR   = InstructionCode("OR")
	ORI  = InstructionCode("ORI")
	AND  = InstructionCode("AND")
	ANDI = InstructionCode("ANDI")
	// If
	SLT   = InstructionCode("SLT")
	SLTI  = InstructionCode("SLTI")
	SLTU  = InstructionCode("SLTU")
	SLTIU = InstructionCode("SLTIU")
	// Branch
	BEQ  = InstructionCode("BEQ")
	BNE  = InstructionCode("BNE")
	BLT  = InstructionCode("BLT")
	BGE  = InstructionCode("BGE")
	BLTU = InstructionCode("BLTU")
	BGEU = InstructionCode("BGEU")
	// Jump
	JAL  = InstructionCode("JAL")
	JALR = InstructionCode("JALR")
	// Synchronize
	FENCE = InstructionCode("FENCE")
	// Environment
	ECALL  = InstructionCode("ECALL")
	EBREAK = InstructionCode("EBREAK")
	// Load
	LB  = InstructionCode("LB")
	LH  = InstructionCode("LH")
	LBU = InstructionCode("LBU")
	LHU = InstructionCode("LHU")
	LW  = InstructionCode("LW")
	// Store
	SB = InstructionCode("SB")
	SH = InstructionCode("SH")
	SW = InstructionCode("SW")

	/*
	 * RV64I
	 */

	// Shift
	SLLW  = InstructionCode("SLLW")
	SLLIW = InstructionCode("SLLIW")
	SRLW  = InstructionCode("SRLW")
	SRLIW = InstructionCode("SRLIW")
	SRAW  = InstructionCode("SRAW")
	SRAIW = InstructionCode("SRAIW")
	// Arithmetic
	ADDW  = InstructionCode("ADDW")
	ADDIW = InstructionCode("ADDIW")
	SUBW  = InstructionCode("SUBW")
	// Load
	LWU = InstructionCode("LWU")
	LD  = InstructionCode("LD")
	// Store
	SD = InstructionCode("SD")

	/*
	 * RV Privileged
	 */

	// Trap
	URET = InstructionCode("URET")
	SRET = InstructionCode("SRET")
	MRET = InstructionCode("MRET")
	// Interrupt
	WFI = InstructionCode("WFI")
	// MMU
	SFENCE_VMA = InstructionCode("SFENCE.VMA")

	/*
	 * Zifencei
	 */
	FENCE_I = InstructionCode("FENCE.I")

	/*
	 * Zicsr
	 */
	CSRRW  = InstructionCode("CSRRW")
	CSRRS  = InstructionCode("CSRRS")
	CSRRC  = InstructionCode("CSRRC")
	CSRRWI = InstructionCode("CSRRWI")
	CSRRSI = InstructionCode("CSRRSI")
	CSRRCI = InstructionCode("CSRRCI")

	/*
	 * RV32A
	 */
	LR_W      = InstructionCode("LR_W")
	SC_W      = InstructionCode("SC_W")
	AMOSWAP_W = InstructionCode("AMOSWAP_W")
	AMOADD_W  = InstructionCode("AMOADD_W")
	AMOXOR_W  = InstructionCode("AMOXOR_W")
	AMOAND_W  = InstructionCode("AMOAND_W")
	AMOOR_W   = InstructionCode("AMOOR_W")
	AMOMIN_W  = InstructionCode("AMOMIN_W")
	AMOMAX_W  = InstructionCode("AMOMAX_W")
	AMOMINU_W = InstructionCode("AMOMINU_W")
	AMOMAXU_W = InstructionCode("AMOMAXU_W")

	/*
	 * RV64A
	 */
	LR_D      = InstructionCode("LR_D")
	SC_D      = InstructionCode("SC_D")
	AMOSWAP_D = InstructionCode("AMOSWAP_D")
	AMOADD_D  = InstructionCode("AMOADD_D")
	AMOXOR_D  = InstructionCode("AMOXOR_D")
	AMOAND_D  = InstructionCode("AMOAND_D")
	AMOOR_D   = InstructionCode("AMOOR_D")
	AMOMIN_D  = InstructionCode("AMOMIN_D")
	AMOMAX_D  = InstructionCode("AMOMAX_D")
	AMOMINU_D = InstructionCode("AMOMINU_D")
	AMOMAXU_D = InstructionCode("AMOMAXU_D")

	/*
	 * RV32M
	 */
	MUL    = InstructionCode("MUL")
	MULH   = InstructionCode("MULH")
	MULHSU = InstructionCode("MULHSU")
	MULHU  = InstructionCode("MULHU")
	DIV    = InstructionCode("DIV")
	DIVU   = InstructionCode("DIVU")
	REM    = InstructionCode("REM")
	REMU   = InstructionCode("REMU")

	/*
	 * RV64M
	 */
	MULW  = InstructionCode("MULW")
	DIVW  = InstructionCode("DIVW")
	DIVUW = InstructionCode("DIVUW")
	REMW  = InstructionCode("REMW")
	REMUW = InstructionCode("REMUW")
)

func (ic InstructionCode) String() string {
	return string(ic)
}

// Instructions is the mapping of InstructionCode and the operation to be executed.
// raw is the instruction machine code, which will be either 32-bit or 16-bit (if compressed).
// pc is the program counter at which the instruction is fetched.
// Note that cpu.PC is not the point the instruction is fetched, because it is already incremented on for the next iteration.
var Instructions = map[InstructionCode]func(cpu *CPU, raw, pc uint64) *Exception{
	/*
	 * RV32C
	 */

	// Load
	C_LW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]

		addr := cpu.rxreg(rs1) + offset
		v, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(v))))
		return ExcpNone()
	},
	C_LWSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		uimm := (bit(raw, 12) << 5) | (bits(raw, 6, 4) << 2) | (bits(raw, 3, 2) << 6)
		v, excp := cpu.Read(cpu.rxreg(2)+uimm, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(v))))
		return ExcpNone()
	},
	C_FLD: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bits(raw, 6, 5) << 6) // raw[6:5] -> offset[7:6]

		v, excp := cpu.Read(cpu.rxreg(rs1)+uint64(offset), doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wfreg(rd, math.Float64frombits(v))
		return ExcpNone()
	},
	C_FLDSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		uimm := (bit(raw, 12) << 5) | (bits(raw, 6, 5) << 3) | (bits(raw, 4, 2) << 6)
		r, excp := cpu.Read(cpu.rxreg(2)+uimm, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		v := math.Float64frombits(r)
		cpu.wfreg(rd, v)
		return ExcpNone()
	},

	// Store
	C_SW: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1 := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 2) | // raw[6] -> offset[2]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		addr := cpu.rxreg(rs1) + offset
		cpu.Write(addr, cpu.rxreg(rs2), word)
		return ExcpNone()
	},
	C_SWSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rs2 := bits(raw, 4, 2)
		uimm := bits(raw, 12, 9)<<2 | bits(raw, 8, 7)<<6
		addr := cpu.rxreg(2) + uimm
		cpu.Write(addr, cpu.rxreg(rs2), word)
		return ExcpNone()
	},
	C_FSD: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 4, 2) + 8
		rs1 := bits(raw, 9, 7) + 8
		offset := (bits(raw, 12, 10) << 3) | // raw[12:10] -> offset[5:3]
			(bit(raw, 6) << 7) | // raw[6] -> offset[7]
			(bit(raw, 5) << 6) // raw[5] -> offset[6]
		addr := cpu.rxreg(rs1) + offset
		cpu.Write(addr, math.Float64bits(cpu.rfreg(rd)), doubleword)
		return ExcpNone()
	},
	C_FSDSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rs2 := bits(raw, 4, 2)
		uimm := bits(raw, 12, 10)<<3 | bits(raw, 9, 7)<<6
		addr := cpu.rxreg(2) + uimm
		v := cpu.rfreg(rs2)
		cpu.Write(addr, math.Float64bits(v), doubleword)
		return ExcpNone()
	},
	// Arithmetic
	C_ADD: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		rs2 := bits(raw, 6, 2)
		cpu.wxreg(rd, cpu.rxreg(rd)+cpu.rxreg(rs2))
		return ExcpNone()
	},
	C_ADDI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		nzimm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (nzimm & 0b10_0000) != 0 {
			// sign-extend
			nzimm = uint64(int64(int32(int16(nzimm | 0b1111_1111_1100_0000))))
		}
		cpu.wxreg(rd, (nzimm + cpu.rxreg(rd)))
		return ExcpNone()
	},
	C_ADDI4SPN: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 4, 2) + 8
		nzuimm := (bits(raw, 12, 11) << 4) | // raw[12:11] -> nzuimm[5:4]
			(bits(raw, 10, 7) << 6) | // raw[10:7] -> nzuimm[9:6]
			(bit(raw, 6) << 2) | // raw[6] -> nzuimm[2]
			(bit(raw, 5) << 3) // raw[5] -> nzuimm[3]

		if nzuimm == 0 {
			return ExcpIllegalInstruction(raw)
		}

		cpu.wxreg(rd, cpu.rxreg(2)+uint64(nzuimm))
		return ExcpNone()
	},
	C_ADDI16SP: func(cpu *CPU, raw, _ uint64) *Exception {
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
			return ExcpNone()
		}

		// write to stack pointer (x2)
		cpu.wxreg(2, cpu.rxreg(2)+imm)
		return ExcpNone()
	},
	C_SUB: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.wxreg(rd, (cpu.rxreg(rd) - cpu.rxreg(rs2)))
		return ExcpNone()
	},
	C_AND: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.wxreg(rd, (cpu.rxreg(rd) & cpu.rxreg(rs2)))
		return ExcpNone()
	},
	C_ANDI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		uimm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (uimm & 0b10_0000) != 0 {
			// sign-extend
			uimm = uint64(int64(int32(int16(uimm | 0b1111_1111_1100_0000))))
		}
		cpu.wxreg(rd, (cpu.rxreg(rd) & uimm))
		return ExcpNone()
	},
	C_OR: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.wxreg(rd, (cpu.rxreg(rd) | cpu.rxreg(rs2)))
		return ExcpNone()
	},
	C_XOR: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		cpu.wxreg(rd, (cpu.rxreg(rd) ^ cpu.rxreg(rs2)))
		return ExcpNone()
	},
	C_MV: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		rs2 := bits(raw, 6, 2)
		cpu.wxreg(rd, cpu.rxreg(rs2))
		return ExcpNone()
	},
	C_LI: func(cpu *CPU, raw, pc uint64) *Exception {
		rd := bits(raw, 11, 7)
		imm := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		if (imm & 0b10_0000) != 0 {
			// sign-extend
			imm = uint64(int64(int32(int16(imm | 0b1111_1111_1100_0000))))
		}

		cpu.wxreg(rd, (imm + cpu.rxreg(0)))
		return ExcpNone()
	},
	C_LUI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		if rd == 2 {
			return ExcpNone()
		}

		imm := (bit(raw, 12) << 17) | // raw[12] -> imm[17]
			(bits(raw, 6, 2) << 12) // raw[6:2] -> imm[16:12]
		if (imm & 0b10_0000_0000_0000_0000) != 0 {
			// sign-extend
			imm = uint64(int64(int32((imm | 0b1111_1111_1111_1100_0000_0000_0000_0000))))
		}
		if imm == 0 {
			return ExcpNone()
		}

		// write to stack pointer (x2)
		cpu.wxreg(rd, imm)
		return ExcpNone()
	},

	// Shift
	C_SLLI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.wxreg(rd, (cpu.rxreg(rd) << shamt))
		return ExcpNone()
	},
	C_SRAI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rd))>>shamt))
		return ExcpNone()
	},
	C_SRLI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		shamt := (bit(raw, 12) << 5) | bits(raw, 6, 2)
		cpu.wxreg(rd, (cpu.rxreg(rd) >> shamt))
		return ExcpNone()
	},

	// Branch
	C_BEQZ: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1 := bits(raw, 9, 7) + 8
		offset := (bit(raw, 12) << 8) |
			(bits(raw, 11, 10) << 3) |
			(bits(raw, 6, 5) << 6) |
			(bits(raw, 4, 3) << 1) |
			(bit(raw, 2) << 5)
		if (offset & 0b1_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1111_1110_0000_0000))))
		}

		if cpu.rxreg(rs1) == 0 {
			cpu.PC = pc + offset
		}

		return ExcpNone()
	},
	C_BNEZ: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1 := bits(raw, 9, 7) + 8
		offset := (bit(raw, 12) << 8) |
			(bits(raw, 11, 10) << 3) |
			(bits(raw, 6, 5) << 6) |
			(bits(raw, 4, 3) << 1) |
			(bit(raw, 2) << 5)
		if (offset & 0b1_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1111_1110_0000_0000))))
		}

		if cpu.rxreg(rs1) != 0 {
			cpu.PC = pc + offset
		}

		return ExcpNone()
	},

	// Jump
	C_J: func(cpu *CPU, raw, pc uint64) *Exception {
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
		if (offset & 0b1000_0000_0000) != 0 {
			// sign-extend
			offset = uint64(int64(int32(int16(offset | 0b1111_0000_0000_0000))))
		}
		cpu.PC = pc + offset
		return ExcpNone()
	},
	C_JR: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1 := bits(raw, 11, 7)
		cpu.PC = cpu.rxreg(rs1)
		return ExcpNone()
	},

	// Jump and Link
	C_JALR: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1 := bits(raw, 11, 7)
		t := pc + 2
		cpu.PC = cpu.rxreg(rs1)
		cpu.wxreg(1, t)
		return ExcpNone()
	},

	// System
	C_NOP: func(cpu *CPU, raw, _ uint64) *Exception {
		// nop does nothing
		return ExcpNone()
	},
	C_EBREAK: func(cpu *CPU, raw, pc uint64) *Exception {
		return ExcpBreakpoint(pc)
	},

	/*
	 * RV64C
	 */

	// Load
	C_LD: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1 := bits(raw, 9, 7) + 8
		rd := bits(raw, 4, 2) + 8
		uimm := (bits(raw, 12, 10) << 3) | (bits(raw, 6, 5) << 6)
		r, excp := cpu.Read(cpu.rxreg(rs1)+uimm, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, r)
		return ExcpNone()
	},
	C_LDSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rd := bits(raw, 11, 7)
		uimm := (bit(raw, 12) << 5) | (bits(raw, 6, 5) << 3) | (bits(raw, 4, 2) << 6)
		r, excp := cpu.Read(cpu.rxreg(2)+uimm, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, r)
		return ExcpNone()
	},

	// Store
	C_SD: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1 := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		uimm := (bits(raw, 12, 10) << 3) | (bits(raw, 6, 5) << 6)
		addr := cpu.rxreg(rs1) + uimm
		cpu.Write(addr, cpu.rxreg(rs2), doubleword)
		return ExcpNone()
	},
	C_SDSP: func(cpu *CPU, raw, _ uint64) *Exception {
		rs2 := bits(raw, 4, 2)
		uimm := (bits(raw, 12, 10) << 3) | (bits(raw, 9, 7) << 6)
		addr := cpu.rxreg(2) + uimm
		cpu.Write(addr, cpu.rxreg(rs2), doubleword)
		return ExcpNone()
	},

	// Arithmetic
	C_ADDW: func(cpu *CPU, raw, pc uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		v := uint64(int64(int32(cpu.rxreg(rd) + cpu.rxreg(rs2))))
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	C_ADDIW: func(cpu *CPU, raw, pc uint64) *Exception {
		rd := bits(raw, 11, 7)
		imm := (bit(raw, 12) << 5) | bits(raw, 6, 2)

		if (imm & 0b10_0000) != 0 {
			imm = uint64(int64(int8(imm | 0b1100_0000)))
		}
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rd)+imm))))
		return ExcpNone()
	},
	C_SUBW: func(cpu *CPU, raw, pc uint64) *Exception {
		rd := bits(raw, 9, 7) + 8
		rs2 := bits(raw, 4, 2) + 8
		v := uint64(int64(int32(cpu.rxreg(rd) - cpu.rxreg(rs2))))
		cpu.wxreg(rd, v)
		return ExcpNone()
	},

	/*
	 * RV32I
	 */

	// Shift
	SLL: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b11_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)<<shamt)
		return ExcpNone()
	},
	SLLI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, shamt := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 25, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)<<shamt)
		return ExcpNone()
	},
	SRL: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shift := cpu.rxreg(rs2) & 0b11_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)>>shift)
		return ExcpNone()
	},
	SRLI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)>>shamt)
		return ExcpNone()
	},
	SRA: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shift := cpu.rxreg(rs2) & 0b111111
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))>>shift))
		return ExcpNone()
	},
	SRAI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))>>shamt))
		return ExcpNone()
	},

	// Arithmetic
	ADD: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)+cpu.rxreg(rs2))
		return ExcpNone()
	},
	ADDI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, imm+cpu.rxreg(rs1))
		return ExcpNone()
	},
	SUB: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)-cpu.rxreg(rs2))
		return ExcpNone()
	},
	LUI: func(cpu *CPU, raw, _ uint64) *Exception {
		imm := uint64(int64(int32(uint32(bits(raw, 31, 12) << 12))))
		rd := bits(raw, 11, 7)

		cpu.wxreg(rd, imm)
		return ExcpNone()
	},
	AUIPC: func(cpu *CPU, raw, pc uint64) *Exception {
		imm := uint64(int64(int32(uint32(bits(raw, 31, 12) << 12))))
		rd := bits(raw, 11, 7)
		cpu.wxreg(rd, pc+imm)
		return ExcpNone()
	},

	// Logical
	XOR: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)^cpu.rxreg(rs2))
		return ExcpNone()
	},
	XORI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)^imm)
		return ExcpNone()
	},
	OR: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)|cpu.rxreg(rs2))
		return ExcpNone()
	},
	ORI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)|imm)
		return ExcpNone()
	},
	AND: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)&cpu.rxreg(rs2))
		return ExcpNone()
	},
	ANDI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)&imm)
		return ExcpNone()
	},

	// If
	SLT: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		var v uint64 = 0
		if int64(cpu.rxreg(rs1)) < int64(cpu.rxreg(rs2)) {
			v = 1
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	SLTI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.rxreg(rs1)) < int64(imm) {
			v = 1
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	SLTU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		var v uint64 = 0
		if cpu.rxreg(rs1) < cpu.rxreg(rs2) {
			v = 1
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	SLTIU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		var v uint64 = 0
		// must compare as two's complement
		if cpu.rxreg(rs1) < imm {
			v = 1
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},

	// Branch
	BEQ: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) == cpu.rxreg(rs2) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},
	BNE: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) != cpu.rxreg(rs2) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},
	BLT: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if int64(cpu.rxreg(rs1)) < int64(cpu.rxreg(rs2)) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},
	BGE: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if int64(cpu.rxreg(rs1)) >= int64(cpu.rxreg(rs2)) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},
	BLTU: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) < cpu.rxreg(rs2) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},
	BGEU: func(cpu *CPU, raw, pc uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) >= cpu.rxreg(rs2) {
			cpu.PC = pc + imm
		}
		return ExcpNone()
	},

	// Jump
	JAL: func(cpu *CPU, raw, pc uint64) *Exception {
		rd, imm := bits(raw, 11, 7), ParseJImm(raw)
		tmp := pc + 4
		cpu.wxreg(rd, tmp)
		cpu.PC = pc + imm
		return ExcpNone()
	},
	JALR: func(cpu *CPU, raw, pc uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		tmp := pc + 4
		target := (cpu.rxreg(rs1) + imm) & ^uint64(1)
		cpu.PC = target
		cpu.wxreg(rd, tmp)
		return ExcpNone()
	},

	// Synchronize
	FENCE: func(cpu *CPU, raw, _ uint64) *Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone()
	},

	// Environment
	ECALL: func(cpu *CPU, raw, _ uint64) *Exception {
		switch cpu.mode {
		case user:
			return ExcpEnvironmentCallFromUmode()
		case supervisor:
			return ExcpEnvironmentCallFromSmode()
		case machine:
			return ExcpEnvironmentCallFromMmode()
		default:
			return ExcpIllegalInstruction(raw)
		}
	},
	EBREAK: func(cpu *CPU, raw, pc uint64) *Exception {
		return ExcpBreakpoint(pc)
	},

	// Load
	LB: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.Read(cpu.rxreg(rs1)+imm, 8)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int8(v))))
		return ExcpNone()
	},
	LH: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.Read(cpu.rxreg(rs1)+imm, 16)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int16(v))))
		return ExcpNone()
	},
	LBU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.Read(cpu.rxreg(rs1)+imm, 8)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	LHU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.Read(cpu.rxreg(rs1)+imm, 16)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, v)
		return ExcpNone()
	},
	LW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.Read(cpu.rxreg(rs1)+imm, 32)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(v))))
		return ExcpNone()
	},

	// Store
	SB: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.Write(addr, cpu.rxreg(rs2), byt)
		return ExcpNone()
	},
	SH: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.Write(addr, cpu.rxreg(rs2), halfword)
		return ExcpNone()
	},
	SW: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.Write(addr, cpu.rxreg(rs2), word)
		return ExcpNone()
	},

	/*
	 * RV64I
	 */

	// Shift
	SLLW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)<<shamt))))
		return ExcpNone()
	},
	SLLIW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)<<shamt))))
		return ExcpNone()
	},
	SRLW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(uint32(cpu.rxreg(rs1))>>shamt))))
		return ExcpNone()
	},
	SRLIW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(uint32(cpu.rxreg(rs1))>>shamt))))
		return ExcpNone()
	},
	SRAW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))>>shamt)))
		return ExcpNone()
	},
	SRAIW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))>>shamt)))
		return ExcpNone()
	},

	// Arithmetic
	ADDW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)+cpu.rxreg(rs2)))))
		return ExcpNone()
	},
	ADDIW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)+imm))))
		return ExcpNone()
	},
	SUBW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)-cpu.rxreg(rs2)))))
		return ExcpNone()
	},

	// Load
	LWU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		addr := cpu.rxreg(rs1) + imm
		r, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, r)
		return ExcpNone()
	},
	LD: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		addr := cpu.rxreg(rs1) + imm
		r, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, r)
		return ExcpNone()
	},

	// Store
	SD: func(cpu *CPU, raw, _ uint64) *Exception {
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.Write(addr, cpu.rxreg(rs2), doubleword)
		return ExcpNone()
	},

	/*
	 * RV Privileged
	 */

	// Trap
	URET: func(cpu *CPU, raw, _ uint64) *Exception {
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

		return ExcpNone()
	},
	SRET: func(cpu *CPU, raw, _ uint64) *Exception {
		// First, set CSRs[SEPC] to program counter.
		cpu.PC = cpu.CSR.Read(CsrSEPC)

		// Then, Modify SSTATUS.

		sstatus := cpu.CSR.Read(CsrSSTATUS)

		// Set CPU mode according to SPP
		switch bit(sstatus, CsrStatusSPP) {
		case 0b0:
			cpu.mode = user
		case 0b1:
			cpu.mode = supervisor

			// MPRV must be set 0 if the mode is not Machine.
			if cpu.mode == supervisor {
				mstatus := cpu.CSR.Read(CsrMSTATUS)
				mstatus = clearBit(mstatus, CsrStatusMPRV)
				cpu.CSR.Write(CsrMSTATUS, mstatus)
				cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
			}
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
		cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)

		return ExcpNone()
	},
	MRET: func(cpu *CPU, raw, _ uint64) *Exception {
		// First, set CSRs[MEPC] to program counter.
		cpu.PC = cpu.CSR.Read(CsrMEPC)

		// Then, Modify MSTATUS.

		mstatus := cpu.CSR.Read(CsrMSTATUS)

		// Set CPU mode according to MPP
		switch bits(mstatus, CsrStatusMPPHi, CsrStatusMPPLo) {
		case 0b00:
			cpu.mode = user
			mstatus = clearBit(mstatus, CsrStatusMPRV)
		case 0b01:
			cpu.mode = supervisor
			mstatus = clearBit(mstatus, CsrStatusMPRV)
		case 0b11:
			cpu.mode = machine
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
		cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)

		return ExcpNone()
	},

	// Interrupt
	WFI: func(cpu *CPU, raw, _ uint64) *Exception {
		cpu.Wfi = true
		return ExcpNone()
	},

	// MMU
	SFENCE_VMA: func(cpu *CPU, raw, _ uint64) *Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone()
	},

	/*
	 * Zifencei
	 */
	FENCE_I: func(cpu *CPU, raw, _ uint64) *Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone()
	},

	/*
	 * Zicsr
	 */
	CSRRW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.CSR.Read(imm)
		v := cpu.rxreg(rs1)
		cpu.CSR.Write(imm, v)
		cpu.wxreg(rd, t)

		if imm == CsrSATP {
			cpu.UpdateAddressingMode(v)
		}

		if imm == CsrMSTATUS || imm == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},
	CSRRS: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.CSR.Read(imm)
		v := t | cpu.rxreg(rs1)
		cpu.CSR.Write(imm, v)
		cpu.wxreg(rd, t)

		if imm == CsrSATP {
			cpu.UpdateAddressingMode(v)
		}

		if imm == CsrMSTATUS || imm == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},
	CSRRC: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.CSR.Read(imm)
		v := t & ^(cpu.rxreg(rs1))
		cpu.CSR.Write(imm, v)
		cpu.wxreg(rd, t)

		if imm == CsrSATP {
			cpu.UpdateAddressingMode(v)
		}

		if imm == CsrMSTATUS || imm == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},
	CSRRWI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, imm, csr := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		csr = csr & 0b111111111111
		cpu.wxreg(rd, cpu.CSR.Read(csr))
		cpu.CSR.Write(csr, imm)

		if csr == CsrSATP {
			cpu.UpdateAddressingMode(imm)
		}

		if csr == CsrMSTATUS || csr == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},
	CSRRSI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.CSR.Read(imm)
		v := t | rs1
		cpu.CSR.Write(imm, v) // RS1 is zimm
		cpu.wxreg(rd, t)

		if imm == CsrSATP {
			cpu.UpdateAddressingMode(v)
		}

		if imm == CsrMSTATUS || imm == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},
	CSRRCI: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.CSR.Read(imm)
		v := t & ^(rs1)
		cpu.CSR.Write(imm, v)
		cpu.wxreg(rd, t)

		if imm == CsrSATP {
			cpu.UpdateAddressingMode(v)
		}

		if imm == CsrMSTATUS || imm == CsrSSTATUS {
			cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
		}

		return ExcpNone()
	},

	/*
	 * RV32A
	 * Note that currently AMO is not implemented atomically because
	 * rv just runs on single goroutine and no out-of-order execution happens.
	 * Also, rl/aq are not handled.
	 */
	LR_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1 := bits(raw, 11, 7), bits(raw, 19, 15)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(t))))
		cpu.reserve(addr)

		return ExcpNone()
	},
	SC_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)

		if cpu.reserved(addr) {
			// SC succeeds.
			cpu.cancel(addr)
			cpu.Write(addr, cpu.rxreg(rs2), word)
			cpu.wxreg(rd, 0)
		} else {
			// SC fails.
			cpu.cancel(addr)
			cpu.wxreg(rd, 1)
		}

		return ExcpNone()
	},
	AMOSWAP_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, cpu.rxreg(rs2), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOADD_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, t+cpu.rxreg(rs2), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOXOR_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, uint64(int64(int32(t)^int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOAND_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, uint64(int64(int32(t)&int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOOR_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, uint64(int64(int32(t)|int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOMIN_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int32(t) < int32(t2) {
			cpu.Write(addr, uint64(int64(int32(t))), word)
		} else {
			cpu.Write(addr, uint64(int64(int32(t2))), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOMAX_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int32(t) < int32(t2) {
			cpu.Write(addr, uint64(int64(int32(t2))), word)
		} else {
			cpu.Write(addr, uint64(int64(int32(t))), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOMINU_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if uint32(t) < uint32(t2) {
			cpu.Write(addr, uint64(uint32(t)), word)
		} else {
			cpu.Write(addr, uint64(uint32(t2)), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},
	AMOMAXU_W: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, word)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if uint32(t) < uint32(t2) {
			cpu.Write(addr, uint64(uint32(t2)), word)
		} else {
			cpu.Write(addr, uint64(uint32(t)), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

		return ExcpNone()
	},

	/*
	 * RV64A
	 */
	LR_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1 := bits(raw, 11, 7), bits(raw, 19, 15)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(t))))
		cpu.reserve(addr)

		return ExcpNone()
	},
	SC_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)

		if cpu.reserved(addr) {
			// SC succeeds.
			cpu.Write(addr, cpu.rxreg(rs2), doubleword)
			cpu.wxreg(rd, 0)
		} else {
			// SC fails.
			cpu.wxreg(rd, 1)
		}

		cpu.cancel(addr)

		return ExcpNone()
	},
	AMOSWAP_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOADD_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, t+cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOXOR_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, t^cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOAND_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, t&cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOOR_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		cpu.Write(addr, t|cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOMIN_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int64(t) < int64(t2) {
			cpu.Write(addr, uint64(int64(t)), doubleword)
		} else {
			cpu.Write(addr, uint64(int64(t2)), doubleword)
		}

		cpu.wxreg(rd, uint64(int64(t)))

		return ExcpNone()
	},
	AMOMAX_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int64(t) < int64(t2) {
			cpu.Write(addr, uint64(int64(t2)), doubleword)
		} else {
			cpu.Write(addr, uint64(int64(t)), doubleword)
		}

		cpu.wxreg(rd, uint64(int64(t)))

		return ExcpNone()
	},
	AMOMINU_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if t < t2 {
			cpu.Write(addr, t, doubleword)
		} else {
			cpu.Write(addr, t2, doubleword)
		}

		cpu.wxreg(rd, t)

		return ExcpNone()
	},
	AMOMAXU_D: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.Read(addr, doubleword)
		if excp.Code != ExcpCodeNone {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if t < t2 {
			cpu.Write(addr, t2, doubleword)
		} else {
			cpu.Write(addr, t, doubleword)
		}

		cpu.wxreg(rd, t)

		return ExcpNone()
	},

	/*
	 * RV32M
	 */
	MUL: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))*int64(cpu.rxreg(rs2))))
		return ExcpNone()
	},
	MULH: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		v1 := cpu.rxreg(rs1)
		v2 := cpu.rxreg(rs2)
		// multiply as signed * signed
		bv1 := big.NewInt(int64(v1))
		bv2 := big.NewInt(int64(v2))
		bv1.Mul(bv1, bv2) // bv1 = bv1 * bv2
		bv1.Rsh(bv1, 64)  // bv1 = bv1 >> 64
		cpu.wxreg(rd, bv1.Uint64())
		return ExcpNone()
	},
	MULHSU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		v1 := cpu.rxreg(rs1)
		v2 := cpu.rxreg(rs2)
		// multiply as signed * unsigned
		var bv1, bv2 big.Int
		bv1.SetInt64(int64(v1))
		bv2.SetUint64(v2)
		bv1.Mul(&bv1, &bv2) // bv1 = bv1 * bv2
		bv1.Rsh(&bv1, 64)   // bv1 = bv1 >> 64
		cpu.wxreg(rd, uint64(bv1.Int64()))
		return ExcpNone()
	},
	MULHU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		v1 := cpu.rxreg(rs1)
		v2 := cpu.rxreg(rs2)
		// multiply as unsigned * unsigned
		var bv1, bv2 big.Int
		bv1.SetUint64(v1)
		bv2.SetUint64(v2)
		bv1.Mul(&bv1, &bv2) // bv1 = bv1 * bv2
		bv1.Rsh(&bv1, 64)   // bv1 = bv1 >> 64
		cpu.wxreg(rd, bv1.Uint64())
		return ExcpNone()
	},
	DIV: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int64(cpu.rxreg(rs1))
		divisor := int64(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.CSR.Read(CsrFCSR)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.CSR.Write(CsrFCSR, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else if dividend == math.MinInt64 && divisor == -1 {
			cpu.wxreg(rd, uint64(dividend))
		} else {
			cpu.wxreg(rd, uint64(dividend/divisor))
		}

		return ExcpNone()
	},
	DIVU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := cpu.rxreg(rs1)
		divisor := cpu.rxreg(rs2)
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.CSR.Read(CsrFCSR)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.CSR.Write(CsrFCSR, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else {
			cpu.wxreg(rd, dividend/divisor)
		}

		return ExcpNone()
	},
	REM: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int64(cpu.rxreg(rs1))
		divisor := int64(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, uint64(dividend))
		} else if dividend == math.MinInt64 && divisor == -1 {
			// overflow. reminder is 0
			cpu.wxreg(rd, 0)
		} else {
			cpu.wxreg(rd, uint64(dividend%divisor))
		}

		return ExcpNone()
	},
	REMU: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := cpu.rxreg(rs1)
		divisor := cpu.rxreg(rs2)
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, dividend)
		} else {
			cpu.wxreg(rd, dividend%divisor)
		}

		return ExcpNone()
	},

	/*
	 * RV64M
	 */
	MULW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))*int32(cpu.rxreg(rs2)))))
		return ExcpNone()
	},
	DIVW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int32(cpu.rxreg(rs1))
		divisor := int32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.CSR.Read(CsrFCSR)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.CSR.Write(CsrFCSR, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else if dividend == math.MinInt32 && divisor == -1 {
			cpu.wxreg(rd, uint64(int64(dividend)))
		} else {
			cpu.wxreg(rd, uint64(int64(dividend/divisor)))
		}

		return ExcpNone()
	},
	DIVUW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := uint32(cpu.rxreg(rs1))
		divisor := uint32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.CSR.Read(CsrFCSR)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.CSR.Write(CsrFCSR, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else {
			cpu.wxreg(rd, uint64(int64(int32(dividend/divisor))))
		}

		return ExcpNone()
	},
	REMW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int32(cpu.rxreg(rs1))
		divisor := int32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, uint64(int64(dividend)))
		} else if dividend == math.MinInt32 && divisor == -1 {
			// overflow. reminder is 0
			cpu.wxreg(rd, 0)
		} else {
			cpu.wxreg(rd, uint64(int64(dividend%divisor)))
		}

		return ExcpNone()
	},
	REMUW: func(cpu *CPU, raw, _ uint64) *Exception {
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := uint32(cpu.rxreg(rs1))
		divisor := uint32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, uint64(int64(int32(dividend))))
		} else {
			cpu.wxreg(rd, uint64(int64(int32(dividend%divisor))))
		}

		return ExcpNone()
	},
}

func ParseIImm(inst uint64) uint64 {
	// inst[31:20] -> immediate[11:0].
	return signExtend(bits(inst, 31, 20), 12)
}

func ParseSImm(inst uint64) uint64 {
	// inst[31:25] -> immediate[11:5], inst[11:7] -> immediate[4:0].
	return signExtend((bits(inst, 11, 7) | bits(inst, 31, 25)<<5), 12)
}

func ParseBImm(inst uint64) uint64 {
	// inst[31:25] -> immediate[12|10:5], inst[11:7] -> immediate[4:1|11].
	return signExtend((bit(inst, 31)<<12)|(bits(inst, 30, 25)<<5)|(bits(inst, 11, 8)<<1)|(bit(inst, 7)<<11), 13)
}

func ParseJImm(inst uint64) uint64 {
	// inst[31:12] -> immediate[20|10:1|11|19:12].
	return signExtend((bit(inst, 31)<<20)|(bits(inst, 30, 21)<<1)|(bit(inst, 20)<<11)|(bits(inst, 19, 12)<<12), 21)
}
