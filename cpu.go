package main

import (
	"fmt"
)

type Bus struct {
	Memory *Memory
	// TODO: Add some more peripheral devices, such as UART.
}

func NewBus() *Bus {
	return &Bus{
		Memory: NewMemory(),
	}
}

func (bus *Bus) Read(addr uint64, size uint8) uint64 {
	return bus.Memory.Read(addr, size)
}

type Mode uint8

const (
	User Mode = iota + 1
	Supervisor
	Machine
)

type CPU struct {
	// program counter
	PC uint64
	// System Bus
	Bus *Bus
	// CPU mode
	Mode Mode

	// Registers
	XRegs *Registers
}

func NewCPU() *CPU {
	return &CPU{
		PC:    0,
		Bus:   NewBus(),
		XRegs: NewRegisters(),
	}
}

func (cpu *CPU) Run() {
	// TODO: eventually physical <-> virtual memory translation must take place here.

	fmt.Printf("[debug] PC: 0x%x\n", cpu.PC)

	// Fetch 32-bit
	inst := cpu.Bus.Read(cpu.PC, 32)
	fmt.Printf("[debug] fetched: %b\n", inst)

	// Decodes the fetched 32-bit instruction.
	// Note that rd, funct3, rs1... does not always match the instruction format,
	// but they are decoded only by its location in the 32-bit.
	opcode := inst & 0x00_00_00_7f
	rd := inst & 0x00_00_0F_80 >> 7
	funct3 := inst & 0x00_00_70_00 >> 12
	rs1 := inst & 0x00_0F_80_00 >> 15
	rs2 := inst & 0x01_F0_00_00 >> 20
	funct7 := inst & 0xFE_00_00_00 >> 25

	// Exec
	switch opcode {
	case 0b011_0011:
		switch funct3 {
		case 0b000:
			switch funct7 {
			case 0b000_0000:
				// add
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)+cpu.XRegs.Read(rs2))
			case 0b010_0000:
				// sub
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)-cpu.XRegs.Read(rs2))
			}
		case 0b001:
			// sll
			// In RV64I, only the low 6 bits of rs2 are used as the shift amount
			shift := cpu.XRegs.Read(rs2) & 0b111111
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)<<shift)
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
			}
		case 0b110:
			// or
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)|cpu.XRegs.Read(rs2))
		case 0b111:
			// and
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)&cpu.XRegs.Read(rs2))
		}
	case 0b110_0111:
		// jalr
		tmp := cpu.PC + 4
		offset := uint64(int64(int32(inst)) >> 20)
		cpu.PC = (((cpu.XRegs.Read(rs1) + offset) >> 1) << 1) - 4 // sub in advance as the PC is incremented later
		cpu.XRegs.Write(rd, tmp)
	case 0b000_0011:
		switch funct3 {
		case 0b000:
			// lb
			offset := uint64(int64(int32(inst)) >> 20)
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, 8)
			cpu.XRegs.Write(rd, uint64(int64(int8(v))))
		case 0b001:
			// lh
			offset := uint64(int64(int32(inst)) >> 20)
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, 16)
			cpu.XRegs.Write(rd, v)
			cpu.XRegs.Write(rd, uint64(int64(int16(v))))
		case 0b010:
			// lw
			offset := uint64(int64(int32(inst)) >> 20)
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, 32)
			cpu.XRegs.Write(rd, uint64(int64(int32(v))))
		case 0b100:
			// lbu
			offset := uint64(int64(int32(inst)) >> 20)
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, 8)
			cpu.XRegs.Write(rd, v)
		case 0b101:
			// lhu
			offset := uint64(int64(int32(inst)) >> 20)
			v := cpu.Bus.Read(cpu.XRegs.Read(rs1)+offset, 16)
			cpu.XRegs.Write(rd, v)
		}
	case 0b001_0011:
		switch funct3 {
		case 0b000:
			// addi
			imm := uint64(int64(int32(inst)) >> 20)
			cpu.XRegs.Write(rd, imm+cpu.XRegs.Read(rs1))
		case 0b010:
			// slti
			imm := uint64(int64(int32(inst)) >> 20)
			var v uint64 = 0
			// must compare as two's complement
			if int64(cpu.XRegs.Read(rs1)) < int64(imm) {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b011:
			// sltiu
			imm := uint64(int64(int32(inst)) >> 20)
			var v uint64 = 0
			// must compare as two's complement
			if cpu.XRegs.Read(rs1) < imm {
				v = 1
			}
			cpu.XRegs.Write(rd, v)
		case 0b100:
			// xori
			imm := uint64(int64(int32(inst)) >> 20)
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)^imm)
		case 0b110:
			// ori
			imm := uint64(int64(int32(inst)) >> 20)
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)|imm)
		case 0b111:
			// andi
			imm := uint64(int64(int32(inst)) >> 20)
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)&imm)
		case 0b001:
			// slli
			shamt := (inst >> 20) & 0b111111
			cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)<<shamt)
		case 0b101:
			switch funct7 {
			case 0b000_0000:
				// srli
				shamt := (inst >> 20) & 0b111111
				cpu.XRegs.Write(rd, cpu.XRegs.Read(rs1)>>shamt)
			case 0b010_0000:
				// srai
				shamt := (inst >> 20) & 0b111111
				cpu.XRegs.Write(rd, uint64(int64(cpu.XRegs.Read(rs1))>>shamt))
			}
		}
	case 0b000_1111:
		switch funct3 {
		case 0b000:
			// fence
			// Do nothing because rv currently does not reorder the instructions for optimizations.
		case 0b001:
			// fence.i
			// Do nothing because rv currently does not reorder the instructions for optimizations.
		}
	case 0b111_0011:
		switch funct3 {
		case 0b000:
			switch rs2 {
			case 0b000:
				// ecall
				switch cpu.Mode {
				case User:
					return ExcpEnvironmentCallFromUmode
				case SuperVisor:
					return ExcpEnvironmentCallFromSmode
				case Machine:
					return ExcpEnvironmentCallFromMmode
				default:
					return ExcpIllegalInstruction
				}
			case eb001:
				// ebreak
				return ExcpBreakpoint
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
		}
	case 0b010_0011:
		fallthrough
	case 0b110_0011:
		fallthrough
	case 0b001_0111:
		// auipc
		imm := uint64(int64(int32(inst & 0xfffff000)))
		cpu.XRegs.Write(rd, cpu.PC+imm)
	case 0b011_0111:
		// lui
		imm := uint64(int64(int32(inst & 0xfffff000)))
		cpu.XRegs.Write(rd, imm)
	case 0b110_1111:
		// jal
		tmp := cpu.PC + 4
		cpu.XRegs.Write(rd, tmp)
		offset := uint64(int64(int32(inst&0x80_00_00_00))) |
			(inst & 0xf_f0_00) |
			((inst >> 9) & 0x8_00) |
			(inst >> 20 & 0x7_fe)
		cpu.PC = offset - 4 // sub in advance as the PC is incremented later
	default:
		return ExcpIllegalInstruction
	}

	cpu.PC += 4
	return nil
}
