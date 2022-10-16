package main

import (
	"fmt"
)

const (
	// mode
	user       = 0
	supervisor = 1
	machine    = 3

	// xlen. 32bit, 128bit aren't supported in rv.
	xlen64 = 1

	// memory size
	byt        = 8
	halfword   = 16
	word       = 32
	doubleword = 64

	// csr stuff
	sstatusmask        = 0b1000000000000000000000000000001100000000000011011110011101100010
	siemask            = 0b1000100010
	sipmask            = 0b1000100010
	ustatus     uint64 = 0x000
	utvec       uint64 = 0x005
	uepc        uint64 = 0x041
	ucause      uint64 = 0x042
	utval       uint64 = 0x043
	fflags      uint64 = 0x001
	frm         uint64 = 0x002
	fcsr        uint64 = 0x003
	sstatus     uint64 = 0x100
	sedeleg     uint64 = 0x102
	sie         uint64 = 0x104
	stvec       uint64 = 0x105
	sepc        uint64 = 0x141
	scause      uint64 = 0x142
	stval       uint64 = 0x143
	sip         uint64 = 0x144
	satp        uint64 = 0x180
	mstatus     uint64 = 0x300
	medeleg     uint64 = 0x302
	mie         uint64 = 0x304
	mtvec       uint64 = 0x305
	mepc        uint64 = 0x341
	mcause      uint64 = 0x342
	mtval       uint64 = 0x343
	mip         uint64 = 0x344

	// memory access type used in address translation
	maInst  = 1
	maLoad  = 2
	maStore = 3

	// addressing mode. sv32, sv48 aren't supported in rv
	amnone = 0
	sv39   = 1
)

type CPU struct {
	// program counter
	PC uint64

	bus *bus

	mode int
	xlen int

	csr [4096]uint64

	xregs [32]uint64
	fregs [32]float64

	lrsc map[uint64]struct{}

	AddressingMode int
	PPN            uint64

	// Wfi represents "wait for interrupt". When this is true, CPU does not run until
	// an interrupt occurs.
	Wfi bool

	// If true, virtual->physical address translation is made.
	PagingEnabled bool
}

func NewCPU() *CPU {
	return &CPU{
		PC: 0,
		bus: &bus{
			ram: NewMemory(),
		},
		mode:           machine,
		csr:            [4096]uint64{},
		xlen:           xlen64,
		xregs:          [32]uint64{},
		fregs:          [32]float64{},
		lrsc:           make(map[uint64]struct{}),
		AddressingMode: 0,
		PPN:            0,
		PagingEnabled:  false,
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
	if addr == fflags {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		return cpu.csr[fcsr] & 0x1f
	}

	if addr == frm {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		return (cpu.csr[fcsr] & 0xe0) >> 5
	}

	if addr == sstatus {
		return cpu.csr[mstatus] & sstatusmask
	}

	if addr == sip {
		return cpu.csr[mip] & sipmask // sip is a subset of mip
	}

	if addr == sie {
		return cpu.csr[mie] & siemask // sie is a subset of mie
	}

	return cpu.csr[addr]
}

func (cpu *CPU) wcsr(addr uint64, value uint64) {
	if addr == fflags {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		cpu.csr[fcsr] &= ^uint64(0x1f) // clear fcsr[4:0]
		cpu.csr[fcsr] |= value & 0x1f  // write the value[4:0] to the fcsr[4:0]
		return
	}

	if addr == frm {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		cpu.csr[fcsr] &= ^uint64(0xe0)       // clear fcsr[7:5]
		cpu.csr[fcsr] |= (value << 5) & 0xe0 // write the value[2:0] to the fcsr[7:5]
		return
	}

	if addr == sstatus {
		// sstatus is a subset of mstatus
		cpu.csr[mstatus] &= ^uint64(sstatusmask) // clear mask
		cpu.csr[mstatus] |= value & sstatusmask  // write only mask
	}

	if addr == sip {
		cpu.csr[mip] &= ^uint64(sipmask)
		cpu.csr[mip] |= value & sipmask
	}

	if addr == sie {
		cpu.csr[mie] &= ^uint64(siemask)
		cpu.csr[mie] |= value & siemask
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
func (cpu *CPU) Fetch(size int) (uint64, *Exception) {
	pAddr, excp := cpu.translate(cpu.PC, maInst, cpu.mode)
	if excp.Code != ExcpCodeNone {
		return 0, excp
	}

	return cpu.bus.Read(pAddr, size), ExcpNone()
}

func (cpu *CPU) Read(addr uint64, size int) (uint64, *Exception) {
	pAddr, excp := cpu.translate(addr, maLoad, cpu.mode)
	if excp.Code != ExcpCodeNone {
		return 0, excp
	}

	return cpu.bus.Read(pAddr, size), ExcpNone()
}

func (cpu *CPU) Write(addr, val uint64, size int) *Exception {
	// Cancel reserved memory to make SC fail when an write is called
	// between LR and SC.
	if cpu.reserved(addr) {
		cpu.cancel(addr)
	}

	pAddr, excp := cpu.translate(addr, maStore, cpu.mode)
	if excp.Code != ExcpCodeNone {
		return excp
	}

	cpu.bus.Write(pAddr, val, size)
	return ExcpNone()
}

func (cpu *CPU) translate(vAddr uint64, ma int, curMode int) (uint64, *Exception) {
	switch cpu.AddressingMode {
	case 0:
		return vAddr, ExcpNone()
	case sv39:
		switch curMode {
		case machine:
			if ma == maInst {
				return vAddr, ExcpNone()
			}

			if ((cpu.rcsr(mstatus) >> 17) & 1) == 0 {
				return vAddr, ExcpNone()
			}

			newPrivMode := (cpu.rcsr(mstatus) >> 9) & 3
			switch newPrivMode {
			case 3:
				return vAddr, ExcpNone()
			default:
				return cpu.translate(vAddr, ma, int(newPrivMode))
			}
		case user, supervisor:
			vpns := []uint64{
				(vAddr >> 12) & 0x1ff,
				(vAddr >> 21) & 0x1ff,
				(vAddr >> 30) & 0x1ff,
			}
			return cpu.TraversePage(vAddr, 3-1, cpu.PPN, vpns, ma)
		default:
			return vAddr, ExcpNone()
		}
	default:
		panic("should not come here")
	}
}

func (cpu *CPU) TraversePage(vAddr uint64, level int, parentPPN uint64, vpns []uint64, ma int) (uint64, *Exception) {
	fault := func() *Exception {
		switch ma {
		case maInst:
			return ExcpInstructionPageFault(vAddr)
		case maLoad:
			return ExcpLoadPageFault(vAddr)
		case maStore:
			return ExcpStoreAMOPageFault(vAddr)
		}

		return ExcpNone() // should not come here
	}

	pageint := 4096

	pteint := 8

	pteAddr := parentPPN*uint64(pageint) + vpns[level]*uint64(pteint)

	pte := cpu.bus.Read(pteAddr, doubleword)

	var ppn uint64
	var ppns []uint64
	if cpu.AddressingMode == sv39 {
		ppn = (pte >> 10) & 0xfffffffffff
		ppns = []uint64{
			(pte >> 10) & 0x1ff,
			(pte >> 19) & 0x1ff,
			(pte >> 28) & 0x3ff_ffff,
		}
	} else {
		panic("unexpected addressing mode!")
	}

	d := (pte >> 7) & 1
	a := (pte >> 6) & 1
	x := (pte >> 3) & 1
	w := (pte >> 2) & 1
	r := (pte >> 1) & 1
	v := pte & 1

	if v == 0 || (r == 0 && w == 1) {
		return 0, fault()
	}

	if r == 0 && x == 0 {
		if level == 0 {
			return 0, fault()
		}

		return cpu.TraversePage(vAddr, level-1, ppn, vpns, ma)
	}

	// page found

	b := false
	if ma == maStore {
		b = d == 0
	}

	if a == 0 || b {
		newPTE := pte | (1 << 6)
		if ma == maStore {
			newPTE |= (1 << 7)
		} else {
			newPTE |= 0
		}

		cpu.bus.Write(pteAddr, newPTE, doubleword)
	}

	switch ma {
	case maInst:
		if x == 0 {
			return 0, fault()
		}
	case maLoad:
		if r == 0 {
			return 0, fault()
		}
	case maStore:
		if w == 0 {
			return 0, fault()
		}
	}

	offset := vAddr & 0xfff
	switch level {
	case 2:
		if ppns[1] != 0 || ppns[0] != 0 {
			return 0, fault()
		}

		return (ppns[2] << 30) | (vpns[1] << 21) | (vpns[0] << 12) | offset, ExcpNone()
	case 1:
		if ppns[0] != 0 {
			return 0, fault()
		}

		return (ppns[2] << 30) | (ppns[1] << 21) | (vpns[0] << 12) | offset, ExcpNone()
	case 0:
		return (ppn << 12) | offset, ExcpNone()
	default:
		panic("invalid level") // should not come here
	}
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

	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	isCompressed := false
	last2bit := raw & 0b11
	if last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10 {
		isCompressed = true
	}

	// As of here, we are not sure if the next instruction is compressed. First we have to figure that out.
	if isCompressed {
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

// Exec executes the decoded instruction.
func (cpu *CPU) Exec(code InstructionCode, raw, cur uint64) *Exception {
	execution, ok := Instructions[code]
	if !ok {
		return ExcpIllegalInstruction(raw)
	}

	return execution(cpu, raw, cur)
}

func (cpu *CPU) UpdateAddressingMode(v uint64) {
	var am int
	switch cpu.xlen {
	case xlen64:
		if v>>60 == 0 {
			am = amnone
		} else if v>>60 == 8 {
			am = sv39
		} else {
			panic("unsupported addressing mode!")
		}
	}

	var ppn uint64
	switch cpu.xlen {
	case xlen64:
		ppn = v & 0xfffffffffff
	}

	cpu.AddressingMode = am
	cpu.PPN = ppn
}

// HandleException catches the raised exception and manipulates CSR and program counter based on
// the exception and CPU privilege mode.
func (cpu *CPU) HandleException(pc uint64, excp *Exception) Trap {
	curPC := pc
	origMode := cpu.mode
	cause := excp.Code

	mdeleg := cpu.rcsr(medeleg)
	sdeleg := cpu.rcsr(sedeleg)

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
		cpu.wcsr(mepc, curPC)

		// MCAUSE is written with a code indicating the event that caused the trap.
		cpu.wcsr(mcause, uint64(cause))

		// MTVAL is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(mtval, excp.TrapValue)

		// PC is updated with the trap-handler base address (MTVEC).
		cpu.PC = cpu.rcsr(mtvec)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if MTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.rcsr(mstatus)
		// update mpie with mie.
		if bit(status, 3) == 0 {
			status = clearBit(status, 7)
		} else {
			status = setBit(status, 7)
		}

		// Clear MIE.
		status = clearBit(status, 3)

		// Update MPP with the previous privilege mode.
		switch origMode {
		case machine:
			status = setBit(status, 11)
			status = setBit(status, 12)
		case supervisor:
			status = setBit(status, 11)
			status = clearBit(status, 12)
		case user:
			status = clearBit(status, 11)
			status = clearBit(status, 12)
		}

		cpu.wcsr(mstatus, status)
	case supervisor:
		// sepc is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.wcsr(sepc, curPC)

		// scause is written with a code indicating the event that caused the trap.
		cpu.wcsr(scause, uint64(cause))

		// stval is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(stval, excp.TrapValue)

		// PC is updated with the trap-handler base address (STVEC).
		cpu.PC = cpu.rcsr(stvec)
		if (cpu.PC & 0b11) != 0 {
			// Add 4 * cause if STVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.PC = (cpu.PC & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
		}

		status := cpu.rcsr(sstatus)
		// update SPIE with SIE.
		if bit(status, 1) == 0 {
			status = clearBit(status, 5)
		} else {
			status = setBit(status, 5)
		}

		// Clear SIE.
		status = clearBit(status, 1)

		// Update SPP with the previous privilege mode.
		switch origMode {
		case supervisor:
			status = setBit(status, 8)
		case user:
			status = clearBit(status, 8)
		}

		cpu.wcsr(sstatus, status)
	case user:
		// uepc is written with the virtual address of the instruction that was
		// interrupted or that encountered the exception.
		cpu.wcsr(uepc, curPC)

		// ucause is written with a code indicating the event that caused the trap.
		cpu.wcsr(ucause, uint64(cause))

		// utval is either set to zero or written with exception-specific information to
		// assist software in handling the trap.
		cpu.wcsr(utval, excp.TrapValue)

		// PC is updated with the trap-handler base address (UTVEC).
		cpu.PC = cpu.rcsr(utvec)
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
