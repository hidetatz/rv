package main

import "fmt"

type Bus struct {
	Memory *Memory
}

func NewBus() *Bus {
	return &Bus{
		Memory: NewMemory(),
	}
}

func (bus *Bus) Read(addr uint64, size Size) uint64 {
	return bus.Memory.Read(addr, size)
}

func (bus *Bus) Write(addr, val uint64, size Size) {
	bus.Memory.Write(addr, val, size)
}

// Mode is RISC-V machine status for privilege architecture.
type Mode uint8

const (
	// User is a mode for application which runs on operating system.
	User Mode = iota + 1
	// Supervisor is a mode for operating system.
	Supervisor
	// Machine is a mode for RISC-V hart internal operation.
	// This sometimes is called kernal-mode or protect-mode in other architecture.
	Machine
)

// XLen is an RISC-V addressing mode.
type XLen uint8

const (
	// XLen64 indicates the 64-bit adressing mode.
	XLen64 = iota + 1

	// 32, 128 are not supported in rv.
)

// Size is the length of memory.
type Size uint8

const (
	Byte       Size = 8
	HalfWord   Size = 16
	Word       Size = 32
	DoubleWord Size = 64
)

type CPU struct {
	// program counter
	PC uint64
	// System Bus
	Bus *Bus
	// CPU mode
	Mode Mode
	XLen XLen

	// Status registers
	CSR CSR

	// Registers
	XRegs *Registers
	FRegs *FRegisters

	// Wfi represents "wait for interrupt". When this is true, CPU does not run until
	// an interrupt occurs.
	Wfi bool
}

func NewCPU() *CPU {
	return &CPU{
		PC:    0,
		Bus:   NewBus(),
		Mode:  Machine,
		CSR:   NewCSR(),
		XLen:  XLen64,
		XRegs: NewRegisters(),
		FRegs: NewFRegisters(),
	}
}

func (cpu *CPU) Fetch(size Size) uint64 {
	return cpu.Bus.Read(cpu.PC, size)
}

func (cpu *CPU) Run() Exception {
	dbg := ""

	if cpu.Wfi {
		return ExcpNone
	}

	// TODO: eventually physical <-> virtual memory translation must take place here.

	// save current PC
	cur := cpu.PC
	Debug("PC: %x", cpu.PC)
	dbg += fmt.Sprintf("  PC: %x", cpu.PC)

	var code InstructionCode

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	raw := cpu.Fetch(HalfWord)

	if IsCompressed(raw) {
		dbg += ", compressed: true"
		dbg += fmt.Sprintf(", raw: %04x", raw)
		code = cpu.DecodeCompressed(raw)
		cpu.PC += 2
	} else {
		dbg += ", compressed: false"
		raw = cpu.Fetch(Word)
		dbg += fmt.Sprintf(", raw: %08x", raw)
		code = cpu.Decode(raw)
		cpu.PC += 4
	}

	if code == _INVALID {
		// TODO: fix
		panic("invalid instruction!")
	}

	dbg += fmt.Sprintf(", Instruction: %s", code)

	excp := cpu.Exec(code, raw, cur)

	dbg += fmt.Sprintf(", SP: %d", cpu.XRegs.Read(2))
	Debug(dbg)

	return excp
}

func (cpu *CPU) Exec(code InstructionCode, raw, cur uint64) Exception {
	execution, ok := Instructions[code]
	if !ok {
		return ExcpIllegalInstruction
	}

	return execution(cpu, raw, cur)
}

// IsCompressed returns if the instruction is compressed 16-bit one.
func IsCompressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}
