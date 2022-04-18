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
	Debug("------Tick------")

	if cpu.Wfi {
		return ExcpNone
	}

	// TODO: eventually physical <-> virtual memory translation must take place here.

	// save current PC
	cur := cpu.PC
	Debug("  PC: %x", cpu.PC)

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	halfword := cpu.Fetch(HalfWord)

	var (
		decoded *Decoded
		excp    Exception
	)

	compressed := cpu.IsCompressed(halfword)
	if compressed {
		Debug("  compressed: true")
		decoded, excp = cpu.DecodeCompressed(halfword)
	} else {
		cpu.PC += 4
		Debug("  compressed: false")
		decoded, excp = cpu.Decode(cpu.Fetch(Word))
	}

	if excp != ExcpNone {
		if excp == ExcpIllegalInstruction {
			// TODO: fix
			panic("invalid instruction!")
		}
		return excp
	}

	Debug("  Instruction: %s", decoded.Code)

	return cpu.Exec(decoded.Code, decoded.Param, cur)
}

func (cpu *CPU) Exec(code InstructionCode, param InstructionParam, addr uint64) Exception {
	switch i := param.(type) {
	case *InstructionR:
		execution, ok := RInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	case *InstructionI:
		execution, ok := IInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	case *InstructionS:
		execution, ok := SInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	case *InstructionB:
		execution, ok := BInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	case *InstructionU:
		execution, ok := UInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	case *InstructionJ:
		execution, ok := JInstructions[code]
		if !ok {
			return ExcpIllegalInstruction
		}

		return execution(cpu, i, addr)
	}

	return ExcpIllegalInstruction
}
