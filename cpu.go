package main

import (
	"fmt"
)

// XLen is an RISC-V addressing mode.
type XLen uint8

const (
	// XLen32 indicates the 32-bit adressing mode.
	XLen32 XLen = iota + 1
	// XLen64 indicates the 64-bit adressing mode.
	XLen64

	// 128-bit is not supported in rv.
)

// Size is the length of memory.
type Size uint8

const (
	Byte       Size = 8
	HalfWord   Size = 16
	Word       Size = 32
	DoubleWord Size = 64
)

// Reservation is a reserved memory emulation module by LR/SC instructions.
// The internal map contains reserved mem address.
// Before implementing multi-core emulation, probably the map value must be changed to
// hart ID, and the map itself must be changed to sync.Map. However, this is
// efficient for now.
type Reservation struct {
	m map[uint64]struct{}
}

// NewReservation returns an initialized reservation.
func NewReservation() *Reservation {
	return &Reservation{m: map[uint64]struct{}{}}
}

// Reserve reserves the given address.
func (r *Reservation) Reserve(addr uint64) {
	r.m[addr] = struct{}{}
}

// IsReserved returns if the given address is reserved.
func (r *Reservation) IsReserved(addr uint64) bool {
	_, ok := r.m[addr]
	return ok
}

// Cancel cancels the reserve.
func (r *Reservation) Cancel(addr uint64) {
	delete(r.m, addr)
}

const (
	// mode
	user       = 0
	supervisor = 1
	machine    = 3
)

// CPU is an processor emulator in rv.
type CPU struct {
	// program counter
	PC uint64
	// Memory management unit
	MMU *MMU
	// CPU mode
	mode int
	XLen XLen

	// Status registers
	CSR *CSR

	// Registers
	xregs [32]uint64
	fregs [32]float64

	// Reservation for LR/SC
	Reservation *Reservation

	// Wfi represents "wait for interrupt". When this is true, CPU does not run until
	// an interrupt occurs.
	Wfi bool

	// If true, virtual->physical address translation is made.
	PagingEnabled bool
}

// NewCPU returns an empty CPU.
// As of the CPU initialized, the memory does not contain any program,
// so it must be loaded before the execution.
func NewCPU(xlen XLen) *CPU {
	return &CPU{
		PC:            0,
		MMU:           NewMMU(xlen),
		mode:          machine,
		CSR:           NewCSR(),
		XLen:          xlen,
		xregs:         [32]uint64{},
		fregs:         [32]float64{},
		Reservation:   NewReservation(),
		PagingEnabled: false,
	}
}

func (cpu *CPU) rxreg(i uint64) uint64 {
	return cpu.xregs[i]
}

func (cpu *CPU) wxreg(i uint64, val uint64) {
	// x0 is always zero, the write should be discarded in that case
	if i != 0 {
		cpu.xregs[i] = val
	}
}

func (cpu *CPU) rfreg(i uint64) float64 {
	return cpu.fregs[i]
}

func (cpu *CPU) wfreg(i uint64, val float64) {
	// f0 is always zero, the write should be discarded in that case
	if i != 0 {
		cpu.fregs[i] = val
	}
}

func (cpu *CPU) Read(addr uint64, size Size) (uint64, *Exception) {
	return cpu.MMU.Read(addr, size, cpu.mode)
}

func (cpu *CPU) Write(addr, val uint64, size Size) *Exception {
	// Cancel reserved memory to make SC fail when an write is called
	// between LR and SC.
	if cpu.Reservation.IsReserved(addr) {
		cpu.Reservation.Cancel(addr)
	}

	return cpu.MMU.Write(addr, val, size, cpu.mode)
}

// Run executes one fetch-decode-exec.
// If instruction execution raised an exception, it also handles it and do some other stuffs.
func (cpu *CPU) Run() Trap {
	if cpu.Wfi {
		return TrapRequested
	}

	// save current PC
	cur := cpu.PC

	var code InstructionCode

	raw, excp := cpu.Fetch(HalfWord)
	if excp.Code != ExcpCodeNone {
		return cpu.HandleException(cur, excp)
	}

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	if IsCompressed(raw) {
		code = DecodeCompressed(raw)
		cpu.PC += 2
	} else {
		raw, excp = cpu.Fetch(Word)
		if excp.Code != ExcpCodeNone {
			return cpu.HandleException(cur, excp)
		}

		code = Decode(raw)
		cpu.PC += 4
	}

	if code == _INVALID {
		return cpu.HandleException(cur, ExcpIllegalInstruction(raw))
	}

	excp = cpu.Exec(code, raw, cur)

	Debug("------")
	Debug(fmt.Sprintf("PC:0x%x	inst:%032b	code:%s	next:0x%x", cur, raw, code, cpu.PC))
	Debug(fmt.Sprintf("x:%v", cpu.xregs))
	Debug(fmt.Sprintf("f:%v", cpu.fregs))
	Debug(fmt.Sprintf("excp:%v", excp.Code))

	if excp.Code != ExcpCodeNone {
		return cpu.HandleException(cur, excp)
	}

	return TrapRequested
}

// Fetch reads the program-counter address of the memory then returns the read binary.
func (cpu *CPU) Fetch(size Size) (uint64, *Exception) {
	return cpu.MMU.Fetch(cpu.PC, size, cpu.mode)
}

// Exec executes the decoded instruction.
func (cpu *CPU) Exec(code InstructionCode, raw, cur uint64) *Exception {
	execution, ok := Instructions[code]
	if !ok {
		return ExcpIllegalInstruction(raw)
	}

	return execution(cpu, raw, cur)
}

// IsCompressed returns if the instruction is compressed 16-bit one.
func IsCompressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}

func (cpu *CPU) UpdateAddressingMode(v uint64) {
	var am AddressingMode
	switch cpu.XLen {
	case XLen32:
		if v&0x8000_0000 == 0 {
			am = AddressingModeNone
		} else {
			am = AddressingModeSV32
		}
	case XLen64:
		if v>>60 == 0 {
			am = AddressingModeNone
		} else if v>>60 == 8 {
			am = AddressingModeSV39
		} else {
			panic("unsupported addressing mode!")
		}
	}

	var ppn uint64
	switch cpu.XLen {
	case XLen32:
		ppn = v & 0x3fffff
	case XLen64:
		ppn = v & 0xfffffffffff
	}

	cpu.MMU.AddressingMode = am
	cpu.MMU.PPN = ppn
}

// HandleException catches the raised exception and manipulates CSR and program counter based on
// the exception and CPU privilege mode.
func (cpu *CPU) HandleException(pc uint64, excp *Exception) Trap {
	curPC := pc
	origMode := cpu.mode
	cause := excp.Code

	mdeleg := cpu.CSR.Read(CsrMEDELEG)
	sdeleg := cpu.CSR.Read(CsrSEDELEG)

	// First, determine the upcoming mode
	if ((mdeleg >> cause) & 1) == 0 {
		cpu.mode = machine
	} else if ((sdeleg >> cause) & 1) == 0 {
		cpu.mode = supervisor
	} else {
		cpu.mode = user
	}

	// Then, start handling exception in the mode
	switch cpu.mode {
	case machine:
		// MEPC is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.CSR.Write(CsrMEPC, curPC)

		// MCAUSE is written with a code indicating the event that caused the trap.
		cpu.CSR.Write(CsrMCAUSE, uint64(cause))

		// MTVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.CSR.Write(CsrMTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (MTVEC).
		cpu.PC = cpu.CSR.Read(CsrMTVEC)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if MTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.CSR.Read(CsrMSTATUS)
		// update MPIE with MIE.
		if bit(status, CsrStatusMIE) == 0 {
			status = clearBit(status, CsrStatusMPIE)
		} else {
			status = setBit(status, CsrStatusMPIE)
		}

		// Clear MIE.
		status = clearBit(status, CsrStatusMIE)

		// Update MPP with the previous privilege mode.
		switch origMode {
		case machine:
			status = setBit(status, CsrStatusMPPLo)
			status = setBit(status, CsrStatusMPPHi)
		case supervisor:
			status = setBit(status, CsrStatusMPPLo)
			status = clearBit(status, CsrStatusMPPHi)
		case user:
			status = clearBit(status, CsrStatusMPPLo)
			status = clearBit(status, CsrStatusMPPHi)
		}

		cpu.CSR.Write(CsrMSTATUS, status)
		cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
	case supervisor:
		// SEPC is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.CSR.Write(CsrSEPC, curPC)

		// SCAUSE is written with a code indicating the event that caused the trap.
		cpu.CSR.Write(CsrSCAUSE, uint64(cause))

		// STVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.CSR.Write(CsrSTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (STVEC).
		cpu.PC = cpu.CSR.Read(CsrSTVEC)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if STVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.CSR.Read(CsrSSTATUS)
		// update SPIE with SIE.
		if bit(status, CsrStatusSIE) == 0 {
			status = clearBit(status, CsrStatusSPIE)
		} else {
			status = setBit(status, CsrStatusSPIE)
		}

		// Clear SIE.
		status = clearBit(status, CsrStatusSIE)

		// Update SPP with the previous privilege mode.
		switch origMode {
		case supervisor:
			status = setBit(status, CsrStatusSPP)
		case user:
			status = clearBit(status, CsrStatusSPP)
		}

		cpu.CSR.Write(CsrSSTATUS, status)
		cpu.MMU.Mstatus = cpu.CSR.Read(CsrMSTATUS)
	case user:
		// UEPC is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.CSR.Write(CsrUEPC, curPC)

		// UCAUSE is written with a code indicating the event that caused the trap.
		cpu.CSR.Write(CsrUCAUSE, uint64(cause))

		// UTVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.CSR.Write(CsrUTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (UTVEC).
		cpu.PC = cpu.CSR.Read(CsrUTVEC)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if UTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}
	}

	switch excp.Code {
	case ExcpCodeInstructionAddressMisalighed:
		return TrapFatal

	case ExcpCodeInstructionAccessFault:
		return TrapFatal

	case ExcpCodeIllegalInstruction:
		return TrapFatal

	case ExcpCodeBreakpoint:
		return TrapRequested

	case ExcpCodeLoadAddressMisaligned:
		return TrapFatal

	case ExcpCodeLoadAccessFault:
		return TrapFatal

	case ExcpCodeStoreAMOAddressMisaligned:
		return TrapFatal

	case ExcpCodeStoreAMOAccessFault:
		return TrapFatal

	case ExcpCodeEnvironmentCallFromUmode:
		return TrapRequested

	case ExcpCodeEnvironmentCallFromSmode:
		return TrapRequested

	case ExcpCodeEnvironmentCallFromMmode:
		return TrapRequested

	case ExcpCodeInstructionPageFault:
		return TrapInvisible

	case ExcpCodeLoadPageFault:
		return TrapInvisible

	case ExcpCodeStoreAMOPageFault:
		return TrapInvisible

	default:
		// must not come here
		panic("ExcpNone is unexpectedly handled")

	}
}
