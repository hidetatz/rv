package main

import (
	"fmt"
)

const (
	// mode
	user       = 0
	supervisor = 1
	machine    = 3

	// xlen
	xlen32 = 1
	xlen64 = 2
	// 128-bit is not supported in rv.

	// memory size
	byt        = 8
	halfword   = 16
	word       = 32
	doubleword = 64
)

type CPU struct {
	// program counter
	PC uint64
	// Memory management unit
	MMU *MMU

	mode int
	xlen int

	csr [4096]uint64

	xregs [32]uint64
	fregs [32]float64

	lrsc map[uint64]struct{}

	// Wfi represents "wait for interrupt". When this is true, CPU does not run until
	// an interrupt occurs.
	Wfi bool

	// If true, virtual->physical address translation is made.
	PagingEnabled bool
}

func NewCPU(xlen int) *CPU {
	return &CPU{
		PC:            0,
		MMU:           NewMMU(xlen),
		mode:          machine,
		csr:           [4096]uint64{},
		xlen:          xlen,
		xregs:         [32]uint64{},
		fregs:         [32]float64{},
		lrsc:          make(map[uint64]struct{}),
		PagingEnabled: false,
	}
}

/*
 * registers
 */
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

/*
 * csr
 */
func (cpu *CPU) rcsr(addr uint64) uint64 {
	if addr == CsrFFLAGS {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		return cpu.csr[CsrFCSR] & 0x1f
	}

	if addr == CsrFRM {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		return (cpu.csr[CsrFCSR] & 0xe0) >> 5
	}

	// when any of SSTATUS, SIP, SIE is requested, masked MSTATUS, MIP, MIE should be returned because they are subsets.
	// See RISC-V Privileged Architecture Spec 4.1
	if addr == CsrSSTATUS {
		return cpu.csr[CsrMSTATUS] & CsrSstatusMask
	}

	if addr == CsrSIP {
		return cpu.csr[CsrMIP] & CsrSipMask
	}

	if addr == CsrSIE {
		return cpu.csr[CsrMIE] & CsrSieMask
	}

	return cpu.csr[addr]
}

func (cpu *CPU) wcsr(addr uint64, value uint64) {
	if addr == CsrFFLAGS {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		cpu.csr[CsrFCSR] &= ^uint64(0x1f) // clear fcsr[4:0]
		cpu.csr[CsrFCSR] |= value & 0x1f  // write the value[4:0] to the fcsr[4:0]
		return
	}

	if addr == CsrFRM {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		cpu.csr[CsrFCSR] &= ^uint64(0xe0)       // clear fcsr[7:5]
		cpu.csr[CsrFCSR] |= (value << 5) & 0xe0 // write the value[2:0] to the fcsr[7:5]
		return
	}

	if addr == CsrSSTATUS {
		// SSTATUS is a subset of MSTATUS
		cpu.csr[CsrMSTATUS] &= ^uint64(CsrSstatusMask) // clear mask
		cpu.csr[CsrMSTATUS] |= value & CsrSstatusMask  // write only mask
	}

	if addr == CsrSIE {
		// SIE is a subset of MIE
		cpu.csr[CsrMIE] &= ^uint64(CsrSieMask)
		cpu.csr[CsrMIE] |= value & CsrSieMask
	}

	if addr == CsrSIP {
		// SIE is a subset of MIE
		cpu.csr[CsrMIP] &= ^uint64(CsrSieMask)
		cpu.csr[CsrMIP] |= value & CsrSieMask
	}

	cpu.csr[addr] = value
}

/*
 * lrsc
 */
func (cpu *CPU) reserve(addr uint64) {
	cpu.lrsc[addr] = struct{}{}
}

func (cpu *CPU) reserved(addr uint64) bool {
	_, ok := cpu.lrsc[addr]
	return ok
}

func (cpu *CPU) cancel(addr uint64) {
	delete(cpu.lrsc, addr)
}

/*
 * memory
 */
func (cpu *CPU) Read(addr uint64, size int) (uint64, *Exception) {
	return cpu.MMU.Read(addr, size, cpu.mode)
}

func (cpu *CPU) Write(addr, val uint64, size int) *Exception {
	// Cancel reserved memory to make SC fail when an write is called
	// between LR and SC.
	if cpu.reserved(addr) {
		cpu.cancel(addr)
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

	raw, excp := cpu.Fetch(halfword)
	if excp.Code != ExcpCodeNone {
		return cpu.HandleException(cur, excp)
	}

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	if IsCompressed(raw) {
		code = DecodeCompressed(raw)
		cpu.PC += 2
	} else {
		raw, excp = cpu.Fetch(word)
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
func (cpu *CPU) Fetch(size int) (uint64, *Exception) {
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
	switch cpu.xlen {
	case xlen32:
		if v&0x8000_0000 == 0 {
			am = AddressingModeNone
		} else {
			am = AddressingModeSV32
		}
	case xlen64:
		if v>>60 == 0 {
			am = AddressingModeNone
		} else if v>>60 == 8 {
			am = AddressingModeSV39
		} else {
			panic("unsupported addressing mode!")
		}
	}

	var ppn uint64
	switch cpu.xlen {
	case xlen32:
		ppn = v & 0x3fffff
	case xlen64:
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

	mdeleg := cpu.rcsr(CsrMEDELEG)
	sdeleg := cpu.rcsr(CsrSEDELEG)

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
		cpu.wcsr(CsrMEPC, curPC)

		// MCAUSE is written with a code indicating the event that caused the trap.
		cpu.wcsr(CsrMCAUSE, uint64(cause))

		// MTVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(CsrMTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (MTVEC).
		cpu.PC = cpu.rcsr(CsrMTVEC)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if MTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.rcsr(CsrMSTATUS)
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

		cpu.wcsr(CsrMSTATUS, status)
		cpu.MMU.Mstatus = cpu.rcsr(CsrMSTATUS)
	case supervisor:
		// SEPC is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.wcsr(CsrSEPC, curPC)

		// SCAUSE is written with a code indicating the event that caused the trap.
		cpu.wcsr(CsrSCAUSE, uint64(cause))

		// STVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(CsrSTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (STVEC).
		cpu.PC = cpu.rcsr(CsrSTVEC)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if STVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.rcsr(CsrSSTATUS)
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

		cpu.wcsr(CsrSSTATUS, status)
		cpu.MMU.Mstatus = cpu.rcsr(CsrMSTATUS)
	case user:
		// UEPC is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.wcsr(CsrUEPC, curPC)

		// UCAUSE is written with a code indicating the event that caused the trap.
		cpu.wcsr(CsrUCAUSE, uint64(cause))

		// UTVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(CsrUTVAL, excp.TrapValue)

		// PC is updated with the trap-handler base address (UTVEC).
		cpu.PC = cpu.rcsr(CsrUTVEC)
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
