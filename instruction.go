package main

type InstructionCode string

const (
	// RV64I

	_INVALID = InstructionCode("_INVALID")
	ADD      = InstructionCode("ADD")
	SUB      = InstructionCode("SUB")
	SLL      = InstructionCode("SLL")
	SLT      = InstructionCode("SLT")
	SLTU     = InstructionCode("SLTU")
	XOR      = InstructionCode("XOR")
	SRL      = InstructionCode("SRL")
	SRA      = InstructionCode("SRA")
	OR       = InstructionCode("OR")
	AND      = InstructionCode("AND")
	JALR     = InstructionCode("JALR")
	LB       = InstructionCode("LB")
	LH       = InstructionCode("LH")
	LW       = InstructionCode("LW")
	LBU      = InstructionCode("LBU")
	LHU      = InstructionCode("LHU")
	ADDI     = InstructionCode("ADDI")
	SLTI     = InstructionCode("SLTI")
	SLTIU    = InstructionCode("SLTIU")
	XORI     = InstructionCode("XORI")
	ORI      = InstructionCode("ORI")
	ANDI     = InstructionCode("ANDI")
	SLLI     = InstructionCode("SLLI")
	SRLI     = InstructionCode("SRLI")
	SRAI     = InstructionCode("SRAI")
	FENCE    = InstructionCode("FENCE")
	FENCE_I  = InstructionCode("FENCE_I")
	ECALL    = InstructionCode("ECALL")
	EBREAK   = InstructionCode("EBREAK")
	CSRRW    = InstructionCode("CSRRW")
	CSRRS    = InstructionCode("CSRRS")
	CSRRC    = InstructionCode("CSRRC")
	CSRR2WI  = InstructionCode("CSRR2WI")
	CSRRSI   = InstructionCode("CSRRSI")
	CSRRCI   = InstructionCode("CSRRCI")
	AUIPC    = InstructionCode("AUIPC")
	LUI      = InstructionCode("LUI")
	JAL      = InstructionCode("JAL")
)

func (ic InstructionCode) String() string {
	return string(ic)
}

var Instructions = map[InstructionCode]func(cpu *CPU, raw uint64) Exception{
	// RV64I
	ADD: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)+cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SUB: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)-cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SLL: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		shamt := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)<<shamt)
		return ExcpNone
	},
	SLT: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		var v uint64 = 0
		if int64(cpu.XRegs.Read(i.Rs1)) < int64(cpu.XRegs.Read(i.Rs2)) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	SLTU: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		var v uint64 = 0
		if cpu.XRegs.Read(i.Rs1) < cpu.XRegs.Read(i.Rs2) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	XOR: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)^cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	SRL: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		shift := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)>>shift)
		return ExcpNone
	},
	SRA: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		shift := cpu.XRegs.Read(i.Rs2) & 0b111111
		cpu.XRegs.Write(i.Rd, uint64(int64(cpu.XRegs.Read(i.Rs1))>>shift))
		return ExcpNone
	},
	OR: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)|cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	AND: func(cpu *CPU, raw uint64) Exception {
		i := ParseR(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)&cpu.XRegs.Read(i.Rs2))
		return ExcpNone
	},
	JALR: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		tmp := cpu.PC + 4
		target := (cpu.XRegs.Read(i.Rs1) + i.Imm) & ^uint64(1)
		cpu.PC = target - 4 // sub in advance as the PC is incremented later
		cpu.XRegs.Write(i.Rd, tmp)
		return ExcpNone
	},
	LB: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 8)
		cpu.XRegs.Write(i.Rd, uint64(int64(int8(v))))
		return ExcpNone
	},
	LH: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 16)
		cpu.XRegs.Write(i.Rd, uint64(int64(int16(v))))
		return ExcpNone
	},
	LW: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 32)
		cpu.XRegs.Write(i.Rd, uint64(int64(int32(v))))
		return ExcpNone
	},
	LBU: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 8)
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	LHU: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		v := cpu.Bus.Read(cpu.XRegs.Read(i.Rs1)+i.Imm, 16)
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	ADDI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, i.Imm+cpu.XRegs.Read(i.Rs1))
		return ExcpNone
	},
	SLTI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.XRegs.Read(i.Rs1)) < int64(i.Imm) {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	SLTIU: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		var v uint64 = 0
		// must compare as two's complement
		if cpu.XRegs.Read(i.Rs1) < i.Imm {
			v = 1
		}
		cpu.XRegs.Write(i.Rd, v)
		return ExcpNone
	},
	XORI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)^i.Imm)
		return ExcpNone
	},
	ORI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)|i.Imm)
		return ExcpNone
	},
	ANDI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)&i.Imm)
		return ExcpNone
	},
	SLLI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)<<shamt)
		return ExcpNone
	},
	SRLI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, cpu.XRegs.Read(i.Rs1)>>shamt)
		return ExcpNone
	},
	SRAI: func(cpu *CPU, raw uint64) Exception {
		i := ParseI(raw)
		shamt := i.Imm & 0b1_1111
		cpu.XRegs.Write(i.Rd, uint64(int64(cpu.XRegs.Read(i.Rs1))>>shamt))
		return ExcpNone
	},
	FENCE: func(cpu *CPU, raw uint64) Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone
	},
	FENCE_I: func(cpu *CPU, raw uint64) Exception {
		// do nothing because rv currently does not apply any optimizations and no fence is needed.
		return ExcpNone
	},
	ECALL: func(cpu *CPU, raw uint64) Exception {
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
		return ExcpNone
	},
	EBREAK: func(cpu *CPU, raw uint64) Exception {
		return ExcpBreakpoint
		return ExcpNone
	},
	CSRRW: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	CSRRS: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	CSRRC: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	CSRR2WI: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	CSRRSI: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	CSRRCI: func(cpu *CPU, raw uint64) Exception {
		return ExcpNone
	},
	AUIPC: func(cpu *CPU, raw uint64) Exception {
		i := ParseU(raw)
		cpu.XRegs.Write(i.Rd, cpu.PC+i.Imm)
		return ExcpNone
	},
	LUI: func(cpu *CPU, raw uint64) Exception {
		i := ParseU(raw)
		cpu.XRegs.Write(i.Rd, i.Imm)
		return ExcpNone
	},
	JAL: func(cpu *CPU, raw uint64) Exception {
		i := ParseJ(raw)
		tmp := cpu.PC + 4
		if i.Rd == 0b0 {
			i.Rd = 1 // x1 if rd is omitted
		}
		cpu.XRegs.Write(i.Rd, tmp)
		cpu.PC += i.Imm - 4 // sub in advance as the PC is incremented later
		Debug("pc in jal: %v", cpu.PC)
		return ExcpNone
	},
}
