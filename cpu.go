package main

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

	// Registers
	XRegs *Registers
	FRegs *FRegisters
}

func NewCPU() *CPU {
	return &CPU{
		PC:    0,
		Bus:   NewBus(),
		Mode:  Machine,
		XLen:  XLen64,
		XRegs: NewRegisters(),
		FRegs: NewFRegisters(),
	}
}

func (cpu *CPU) Fetch(size Size) uint64 {
	return cpu.Bus.Read(cpu.PC, size)
}

func (cpu *CPU) Run() Exception {
	Debug("Tick-------")
	// TODO: eventually physical <-> virtual memory translation must take place here.

	var inst uint64

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	halfword := cpu.Fetch(HalfWord)
	compressed := cpu.IsCompressed(halfword)
	if compressed {
		// if compressed, extract it to be a 32-bit one.
		decompressed, excp := cpu.Decompress(halfword)
		if excp != ExcpNone {
			return excp
		}
		inst = decompressed
	} else {
		inst = cpu.Fetch(Word)
	}

	Debug("compressed: %v", compressed)
	Debug("inst: %032b", inst)

	// Decode the instruction
	instructionCode := cpu.Decode(inst)

	// Execute the instruction.
	exception := cpu.Exec(instructionCode, inst)
	if compressed {
		cpu.PC += 2
	} else {
		cpu.PC += 4
	}

	return exception
}

func (cpu *CPU) Exec(code InstructionCode, inst uint64) Exception {
	execution, ok := Instructions[code]
	if !ok {
		return ExcpIllegalInstruction
	}

	return execution(cpu, inst)
}
