package main

import (
	"fmt"
	"math"
	"math/big"
)

const (
	modeUser       = 0
	modeSupervisor = 1
	modeHypervisor = 2
	modeMachine    = 3

	xlen32 = 0
	xlen64 = 1

	mByte        = 8
	mHalfword   = 16
	mWord       = 32
	mDoubleword = 64

	// csr stuff
	sstatusmask        = 0b1000000000000000000000000000001100000000000011011110011101100010
	siemask            = 0b1000100010
	sipmask            = 0b1000100010
	ustatus     uint64 = 0x000
	uie         uint64 = 0x004
	utvec       uint64 = 0x005
	uepc        uint64 = 0x041
	ucause      uint64 = 0x042
	utval       uint64 = 0x043
	fflags      uint64 = 0x001
	frm         uint64 = 0x002
	fcsr        uint64 = 0x003
	sstatus     uint64 = 0x100
	sedeleg     uint64 = 0x102
	sideleg     uint64 = 0x103
	sie         uint64 = 0x104
	stvec       uint64 = 0x105
	sepc        uint64 = 0x141
	scause      uint64 = 0x142
	stval       uint64 = 0x143
	sip         uint64 = 0x144
	satp        uint64 = 0x180
	mstatus     uint64 = 0x300
	medeleg     uint64 = 0x302
	mideleg     uint64 = 0x303
	mie         uint64 = 0x304
	mtvec       uint64 = 0x305
	mepc        uint64 = 0x341
	mcause      uint64 = 0x342
	mtval       uint64 = 0x343
	mip         uint64 = 0x344
	cycle       uint64 = 0xc00

	// memory access type used in address translation
	maInst  = 1
	maLoad  = 2
	maStore = 3

	// addressing mode. sv57 isn't supported in rv
	svnone = 0
	sv32   = 1
	sv39   = 2
	sv48   = 3

	// trap

	// exception
	eInstAddrMisaligned  = 0
	eInstAccessFault     = 1
	eIllegalInst         = 2
	eBreakpoint          = 3
	eLoadAddrMisaligned  = 4
	eLoadAccessFault     = 5
	eStoreAddrMisaligned = 6
	eStoreAccessFault    = 7
	eEcallFromU          = 8
	eEcallFromS          = 9
	eEcallFromM          = 11
	eInstPageFault       = 12
	eLoadPageFault       = 13
	eStorePageFault      = 15

	// interrupt
	iUserSoftware       = 0
	iSupervisorSoftware = 1
	iMachineSoftware    = 3
	iUserTimer          = 4
	iSupervisorTimer    = 5
	iMachineTimer       = 7
	iUserExternal       = 8
	iSupervisorExternal = 9
	iMachineExternal    = 11

	// memory management
	drambase = 0x8000_0000
	dtbsize  = 0xfe0
)

type trap struct {
	code  int
	value uint64
}

type CPU struct {
	cycle uint64
	pc uint64
	wfi bool
	xlen int
	privilege int
	x [32]int64
	f [32]float64
	csr *CSR
	mmu *MMU
	testmode bool
}

func NewCPU(machine Machine, console Console, testmode bool) *CPU {
	cpu := &CPU{
		cycle: 0,
		pc: 0,
		wfi: false,
		xlen: xlen64,
		privilege: Machine,
		x: [32]int64{},
		f: [32]float64{},
		csr: NewCSR(),
		mmu: NewMMU(xlen64, machine, console),
		testmode: testmode,
	}

	cpu.x[0xb] = cpu.mmu.getBus().getBaseAddress(device.DTB)

	return cpu
}

func (cpu *CPU) reset() {
	cpu.pc = 0
	cpu.cycle = 0
	cpu.privilege = Machine
	cpu.wfi = false
	cpu.xlen = xlen64
	cpu.x = [32]int64{}
	cpu.f = [32]float64{}
}

func (cpu *CPU) setPC(pc uint64) {
	cpu.pc = pc
}

func (cpu *CPU) setXlen(xlen int) {
	cpu.xlen = xlen
	cpu.mmu.setXlen(xlen)
}

func (cpu *CPU) tick() {
	if intr := cpu.checkIntrrupts(); intr != nil {
		cpu.intrruptHandler(intr)
	}

	if !cpu.wfi {
		instructionAddr = cpu.pc
		if excp := cpu.tickExecute(); excp != nil {
			cpu.catchException(excp, instructionAddr)
		}
	}

	bus := cpu.mmu.getBus()
	irqs = bus.tick()

	cpu.tickIntrrupt(irqs)

	cpu.cycle++
	cpu.csr.writeDirect(CsrCycle, cpu.cycle)
	cpu.csr.tick()
}

func (cpu *CPU) tickExecute() *Trap {
	instructionAddr := cpu.pc
	word, err := cpu.fetch()
	if err != nil {
		return err
	}

	opecode, err := cpu.decode(word)
	if err != nil {
		return err
	}

	instruction, err := opecode.operation(cpu, instructionAddr, word)
	if err != nil {
		panic("instruction not defined: %b", word)
	}

	if err := instruction.operation(instructionAddr, word); err != nil {
		return err
	}

	cpu.x[0] = 0

	return nil
}

func (cpu *CPU) tickIntrrupt(irqs []bool) {
	bus := cpu.mmu.getBus()

	if irqs[machine] {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpMeip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpMeip)
	}

	if irqs[hypervisor] {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpHeip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpHeip)
	}

	if irqs[supervisor] {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpSeip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpSeip)
	}

	if irqs[user] {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpUeip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpUeip)
	}

	if bus.isPendingTimerInterrupt(0) {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpMtip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpMtip)
	}

	if bus.isPendingSoftwareInterrupt(0) {
		cpu.csr.readModifyWriteDirect(CsrMIP, CsrIpMsip, 0)
	} else {
		cpu.csr.readModifyWriteDirect(CsrMIP, 0, CsrIpMsip)
	}
}

func (cpu *CPU) fetch() (uint32, *Trap) {
	fetchWord, err := cpu.mmu.fetch32(cpu.pc)
	if err != nil {
		return 0, err
	}

	if (fetchWord & 0x3) == 0x3 {
		pc += 4
		return fetchWord, nil
	} else {
		pc += 2
		if word, err := cpu.instructionDecompress(pc - 2, fetchWord); err != nil {
			return word, nil
		} else {
			return 0, &Trap{illegalInst, pc - 2}
		}
	}
}

func (cpu *CPU) decode() (Opecode, *Trap) {
	if o := Opecodes.get(word & 0x7f) != nil {
		return o
	}

	panic("opecode not found")
}

func (cpu *CPU) catchException(trap *Trap, addr uint64) {
	trapcode := trap.exception
	previousPrivilege = cpu.privilege
	nextPrivilege := cpu.getNextPrivilege(trapcode, false)
	cpu.changePrivilege(nextPrivilege)
	cpu.updateCsrTrapRegisters(addr, trapcode, trap.value, previousPrivilege, false)
	cpu.pc = cpu.getTrapNextPC()
}

func (cpu *CPU) checkInterrupts() int {
	mie := cpu.csr.readDirect(CsrMIE)
	mip := cpu.csr.readDirect(CsrMIP)
	cause := mie & mip & 0xfff

	if cause & CsrIpMeip > 0 && cpu.selectHandlingInterrupt(iMachineExternal) {
		return iMachineExternal
	}

	if cause & CsrIpMsip > 0 && cpu.selectHandlingInterrupt(iMachineSoftware) {
		return iMachineSoftware
	}

	if cause & CsrIpMtip > 0 && cpu.selectHandlingInterrupt(iMachineTimer) {
		return iMachineTimer
	}

	if cause & CsrIpHeip > 0 {
		panic("unexpected event happened")
	}

	if cause & CsrIpHtip > 0 {
		panic("unexpected event happened")
	}

	if cause & CsrIpHsip > 0 {
		panic("unexpected event happened")
	}

	if cause & CsrIpSeip > 0 && cpu.selectHandlingInterrupt(iSupervisorExternal) {
		return iSupervisorExternal
	}

	if cause & CsrIpSsip > 0 && cpu.selectHandlingInterrupt(iSupervisorSoftware) {
		return iSupervisorSoftware
	}

	if cause & CsrIpStip > 0 && cpu.selectHandlingInterrupt(iSupervisorTimer) {
		return iSupervisorTimer
	}

	if cause & CsrIpUeip > 0 && cpu.selectHandlingInterrupt(iUserExternal) {
		return iUserExternal
	}

	if cause & CsrIpUtip > 0 && cpu.selectHandlingInterrupt(iUserTimer) {
		return iUserTimer
	}

	if cause & CsrIpUsip > 0 && cpu.selectHandlingInterrupt(iUserSoftware) {
		return iUserSoftware
	}

	return nil
}

func (cpu *CPU) interruptHandler(intr int) {
	trapCode := uint8(intr)
	previousPrivilege = cpu.privilege
	nextPrivilege = cpu.getNextPrivilege(trapCode, true)

	cpu.changePrivilege(nextPrivilege)
	cpu.updateCsrTrapRegisters(cpu.pc, trapCode, cpu.pc, previousPrivilege, true)
	cpu.pc = cpu.getTrapNextPC()

	cpu.wfi = false
}

func (cpu *CPU) clearInterrupt(intr int) {
	mip := cpu.csr.readDirect(CsrMip)
	n := 0;
	switch intr {
	case iMachineExternal:
		n = 0x800
	case iSupervisorExternal:
		n = 0x200	
	case iUserExternal:
		n = 0x100
	case iMachineTimer:
		n = 0x080
	case iSupervisorTimer:
		n = 0x020
	case iUserTimer:
		n = 0x010
	case iMachineSoftware:
		n = 0x008
	case iSupervisorSoftware:
		n = 0x002
	case iUserSoftware:
		n = 0x001
	}

	cpu.csr.writeDirect(CsrMip, mip & (^n))
}

func (cpu *CPU) selectHandlingInterrupt(intr int) bool {
	trapCode := uint8(intr)
	nextPrivilege := cpu.getNextPrivilege(trapCode, true)
	ie, status := 0, 0
	switch nextPrivilege {
	case modeUser:
		ie = cpu.csr.readDirect(CsrUie)
		status = cpu.csr.readDirect(CsrUstatus)
	case modeSupervisor:
		ie = cpu.csr.readDirect(CsrSie)
		status = cpu.csr.readDirect(CsrSstatus)
	case modeHypervisor:
		ie = cpu.csr.readDirect(CsrHie)
		status = cpu.csr.readDirect(CsrHstatus)
	case modeMachine:
		ie = cpu.csr.readDirect(CsrMie)
		status = cpu.csr.readDirect(CsrMstatus)
	}

	nextPrivilegeLevel := nextPrivilege
	privilegeLevel := cpu.privilege
	if privilegeLevel < nextPrivilegeLevel {
		return false
	}

	uie := status & 1
	sie := (status >> 1) & 1
	hie := (status >> 2) & 1
	uie := (status >> 3) & 1
	if privilegeLevel == nextPrivilegeLevel {
		switch cpu.privilege {
		case modeUser:
			if uie == 0 {
				return false
			}
		case modeSupervisor:
			if sie == 0 {
				return false
			}
		case modeHypervisor:
			if hie == 0 {
				return false
			}
		case modeMachine:
			if mie == 0 {
				return false
			}
		}
	}

	switch intr {
	case iMachineExternal:
		meie := (ie >> 11) & 1
		if meie == 0 {
			return false
		}
	case iMachineSoftware:
		msie := (ie >> 3) & 1
		if msie == 0 {
			return false
		}
	case iMachineTimer:
		mtie := (ie >> 7) & 1
		if mtie == 0 {
			return false
		}
	case iSupervisorExternal:
		seie := (ie >> 9) & 1
		if seie == 0 {
			return false
		}
	case iSupervisorSoftware:
		ssie := (ie >> 1) & 1
		if ssie == 0 {
			return false
		}
	case iSupervisorTimer:
		stie := (ie >> 5) & 1
		if stie == 0 {
			return false
		}
	case iUserExternal:
		ueie := (ie >> 8) & 1
		if ueie == 0 {
			return false
		}
	case iUserSoftware:
		usie := ie & 1
		if usie == 0 {
			return false
		}
	case iUserTimer:
		utie := (ie >> 4) & 1
		if utie == 0 {
			return false
		}
	}

	return true
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
		return (cpu.csr[fcsr] >> 5) & 0x7
	}

	if addr == sstatus {
		return cpu.csr[mstatus] & 0x80000003000de162
	}

	if addr == sip {
		return cpu.csr[mip] & 0x222 // sip is a subset of mip
	}

	if addr == sie {
		return cpu.csr[mie] & 0x222 // sie is a subset of mie
	}

	return cpu.csr[addr]
}

func (cpu *CPU) wcsr(addr uint64, value uint64) {
	if addr == fflags {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		cpu.csr[fcsr] &= ^uint64(0x1f) // clear fcsr[4:0]
		cpu.csr[fcsr] |= value & 0x1f  // write the value[4:0] to the fcsr[4:0]
	}

	if addr == frm {
		// fcsr consists of frm (3-bit) + fflags (5-bit)
		cpu.csr[fcsr] &= ^uint64(0xe0)       // clear fcsr[7:5]
		cpu.csr[fcsr] |= (value << 5) & 0xe0 // write the value[2:0] to the fcsr[7:5]
	}

	if addr == sstatus {
		// sstatus is a subset of mstatus
		cpu.csr[mstatus] &= ^uint64(0x80000003000de162) // clear mask
		cpu.csr[mstatus] |= value & 0x80000003000de162  // write only mask
	}

	if addr == sip {
		cpu.csr[mip] &= ^uint64(0x222)
		cpu.csr[mip] |= value & 0x222
	}

	if addr == sie {
		cpu.csr[mie] &= ^uint64(0x222)
		cpu.csr[mie] |= value & 0x222
	}

	if addr == mideleg {
		cpu.csr[addr] = value & 0x666
	}

	if addr != 0 {
		cpu.csr[addr] = value
	}

	if addr == satp {
		cpu.updateAddressingMode(value)
	}
}

func (cpu *CPU) updateAddressingMode(value uint64) {
	switch cpu.xlen {
	case xlen32:
		if value&0x80000000 == 0 {
			cpu.addressingMode = svnone
		} else {
			cpu.addressingMode = sv32
		}

		cpu.ppn = value & 0x3fffff
	case xlen64:
		switch value >> 60 {
		case 0:
			cpu.addressingMode = svnone
		case 8:
			cpu.addressingMode = sv39
		case 9:
			cpu.addressingMode = sv48
		default:
			panic("unknown addressing mode")
		}

		cpu.ppn = value & 0xfffffffffff
	}
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

func (cpu *CPU) getEffectiveAddr(addr uint64) uint64 {
	if cpu.xlen == xlen32 {
		return addr & 0xffffffff
	}

	return addr
}

func (cpu *CPU) fetch() (uint64, *trap) {
	vAddr := cpu.pc
	if (vAddr & 0xfff) <= 0x1000-4 {
		eAddr := cpu.getEffectiveAddr(vAddr)
		pa, excp := cpu.translate(eAddr, maInst)
		if excp != nil {
			return 0, &trap{code: instPageFault, value: vAddr}
		}

		v := cpu.readRaw(pa, word)

		return v, nil
	}

	data := uint64(0)
	for i := uint64(0); i < 4; i++ {
		eAddr := cpu.getEffectiveAddr(vAddr + 1)
		pa, excp := cpu.translate(eAddr, maInst)
		if excp != nil {
			return 0, &trap{code: instPageFault, value: vAddr}
		}

		v := cpu.ram.Read(pa, byt)
		data |= v << (i * 8)
	}

	return data, nil
}

func (cpu *CPU) read(vaddr uint64, size int) (uint64, *trap) {
	//if (vaddr & 0xfff) <= 0x1000-uint64(size/8) {
	//	paddr, excp := cpu.translate(vaddr, maLoad)
	//	if excp != nil {
	//		return 0, &trap{code: loadPageFault, value: vaddr}
	//	}

	//	return cpu.readRaw(paddr, size), nil
	//}

	data := uint64(0)
	for i := 0; i < size/8; i++ {
		eaddr := cpu.getEffectiveAddr(vaddr + uint64(i))
		paddr, excp := cpu.translate(eaddr, maLoad)
		if excp != nil {
			return 0, &trap{code: loadPageFault, value: vaddr}
		}

		v := cpu.readRaw(paddr, byt)
		data |= v << (i * 8)
	}

	return data, nil
}

func (cpu *CPU) readRaw(paddr uint64, size int) uint64 {
	eaddr := cpu.getEffectiveAddr(paddr)

	// overflow := false
	// if size > byt {
	// 	overflow = eaddr+uint64(size/8-1) > eaddr
	// }

	// if eaddr >= drambase && !overflow {
	if eaddr >= drambase {
		return cpu.ram.Read(eaddr, size)
	}

	data := uint64(0)
	for i := 0; i < size/8; i++ {
		a := eaddr + uint64(i)
		var d uint8 = 0
		switch {
		case 0x00001020 <= a && a < 0x00001fff:
			d = cpu.dtb[a-0x1020]
		case 0x02000000 <= a && a < 0x0200ffff:
			d = cpu.clint.read(a)
		case 0x0c000000 <= a && a < 0x0fffffff:
			d = cpu.plic.read(a)
		case 0x10000000 <= a && a < 0x100000ff:
			d = cpu.uart.read(a)
		case 0x10001000 <= a && a < 0x10001fff:
			d = cpu.disk.read(a)
		default:
			panic(fmt.Sprintf("unknown mem seg: %b", a))
		}
		data |= uint64(d << (uint8(i) * 8))
	}

	return data
}

func (cpu *CPU) write(vaddr, val uint64, size int) *trap {
	// Cancel reserved memory to make SC fail when an write is called
	// between LR and SC.
	//if cpu.reserved(addr) {
	//	cpu.cancel(addr)
	//}

	//if (addr & 0xfff) <= 0x1000-uint64(size/8) {
	//	eAddr := cpu.getEffectiveAddr(addr)
	//	pAddr, excp := cpu.translate(eAddr, maStore)
	//	if excp != nil {
	//		return &trap{code: storePageFault, value: addr}
	//	}

	//	cpu.ram.Write(pAddr, val, size)
	//}

	for i := 0; i < size/8; i++ {
		a := vaddr + uint64(i)
		v := (val >> (i * 8)) & 0xff
		paddr, excp := cpu.translate(a, maStore)
		if excp != nil {
			return &trap{code: storePageFault, value: a}
		}

		cpu.writeRaw(paddr, v, byt)
	}

	return nil
}

func (cpu *CPU) writeRaw(addr, val uint64, size int) {
	ea := cpu.getEffectiveAddr(addr)

	// overflow := false
	// if size > byt {
	// 	overflow = ea+uint64(size/8-1) > ea
	// }

	// if ea >= drambase && !overflow {
	if ea >= drambase {
		cpu.ram.Write(ea, val, size)
		return
	}

	for i := 0; i < size/8; i++ {
		v := uint8((val >> (i * 8)) & 0xff)
		a := ea + uint64(i)
		switch {
		case 0x02000000 <= a && a < 0x0200ffff:
			cpu.clint.write(a, v)
		case 0x0c000000 <= a && a < 0x0fffffff:
			cpu.plic.write(a, v)
		case 0x10000000 <= a && a < 0x10000fff:
			cpu.uart.write(a, v)
		case 0x10001000 <= a && a < 0x10001fff:
			cpu.disk.write(a, v)
		default:
			panic("unknown mem seg")
		}
	}
}

func (cpu *CPU) translate(vAddr uint64, ma int) (uint64, *trap) {
	eAddr := cpu.getEffectiveAddr(vAddr)

	switch cpu.addressingMode {
	case svnone:
		return eAddr, nil

	case sv32:
		switch cpu.mode {
		case machine:
			switch ma {
			case maInst:
				return eAddr, nil
			default:
				mst := cpu.rcsr(mstatus)
				if (mst>>17)&1 == 0 {
					return eAddr, nil
				}

				newMode := (mst >> 9) & 3
				if newMode == machine {
					return eAddr, nil
				}

				curMode := cpu.mode
				cpu.mode = int(newMode)
				r, excp := cpu.translate(vAddr, ma)
				if excp != nil {
					return 0, excp
				}
				cpu.mode = curMode
				return r, nil

			}
		case supervisor, user:
			vpns := []uint64{(eAddr >> 12) & 0x1ff, (eAddr >> 22) & 0x3ff}
			return cpu.traversePage(eAddr, 2-1, cpu.ppn, vpns, ma)
		}
	case sv39:
		switch cpu.mode {
		case machine:
			switch ma {
			case maInst:
				return eAddr, nil
			default:
				mst := cpu.rcsr(mstatus)
				if (mst>>17)&1 == 0 {
					return eAddr, nil
				}

				newMode := (mst >> 9) & 3
				if newMode == machine {
					return eAddr, nil
				}

				curMode := cpu.mode
				cpu.mode = int(newMode)
				r, excp := cpu.translate(vAddr, ma)
				if excp != nil {
					return 0, excp
				}
				cpu.mode = curMode
				return r, nil

			}
		case supervisor, user:
			vpns := []uint64{(eAddr >> 12) & 0x1ff, (eAddr >> 21) & 0x1ff, (eAddr >> 30) & 0x1ff}
			return cpu.traversePage(eAddr, 3-1, cpu.ppn, vpns, ma)
		}
	case sv48:
		panic("sv48 is unsupported")
	}

	panic("unknown addressing mode")
}

func (cpu *CPU) traversePage(vAddr uint64, level int, parentPPN uint64, vpns []uint64, ma int) (uint64, *trap) {
	fault := func() *trap {
		switch ma {
		case maInst:
			return &trap{code: instPageFault, value: vAddr}
		case maLoad:
			return &trap{code: loadPageFault, value: vAddr}
		case maStore:
			return &trap{code: storePageFault, value: vAddr}
		}

		return nil // should not come here
	}

	pageSize := uint64(4096)
	pteSize := uint64(8)
	if cpu.addressingMode == sv32 {
		pteSize = uint64(4)
	}

	pteAddr := parentPPN*pageSize + vpns[level]*pteSize
	var pte uint64
	if cpu.addressingMode == sv32 {
		pte = cpu.ram.Read(pteAddr, word)
	} else {
		pte = cpu.ram.Read(pteAddr, doubleword)
	}

	var ppn uint64
	if cpu.addressingMode == sv32 {
		ppn = (pte >> 10) & 0x3fffff
	} else {
		ppn = (pte >> 10) & 0xfffffffffff
	}

	var ppns []uint64
	if cpu.addressingMode == sv32 {
		ppns = []uint64{(pte >> 10) & 0x3ff, (pte >> 20) & 0xfff, 0}
	} else {
		ppns = []uint64{(pte >> 10) & 0x1ff, (pte >> 19) & 0x1ff, (pte >> 28) & 0x3ffffff}
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

		return cpu.traversePage(vAddr, level-1, ppn, vpns, ma)
	}

	// page found

	if a == 0 || (ma == maStore && d == 0) {
		newPTE := pte | (1 << 6)
		if ma == maStore {
			newPTE |= (1 << 7)
		}

		if cpu.addressingMode == sv32 {
			cpu.ram.Write(pteAddr, newPTE, word)
		} else {
			cpu.ram.Write(pteAddr, newPTE, doubleword)
		}
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
	switch cpu.addressingMode {
	case sv32:
		switch level {
		case 1:
			if ppns[0] != 0 {
				return 0, fault()
			}

			return (ppns[1] << 22) | (vpns[0] << 12) | offset, nil
		case 0:
			return (ppn << 12) | offset, nil
		default:
			panic("invalid level") // should not come here
		}
	default:
		switch level {
		case 2:
			if ppns[1] != 0 || ppns[0] != 0 {
				return 0, fault()
			}

			return (ppns[2] << 30) | (vpns[1] << 21) | (vpns[0] << 12) | offset, nil
		case 1:
			if ppns[0] != 0 {
				return 0, fault()
			}

			return (ppns[2] << 30) | (ppns[1] << 21) | (vpns[0] << 12) | offset, nil
		case 0:
			return (ppn << 12) | offset, nil
		default:
			panic("invalid level") // should not come here
		}
	}
}

func (cpu *CPU) decompress(inst uint64) uint64 {
	op := inst & 0x3
	funct3 := (inst >> 13) & 0x7

	switch op {
	case 0:
		switch funct3 {
		case 0:
			// C.ADDI4SPN addi rd+8, x2, nzuimm
			rd := (inst >> 2) & 0x7 // [4:2]
			nzuimm :=
				((inst >> 7) & 0x30) | // nzuimm[5:4] <= [12:11]
					((inst >> 1) & 0x3c0) | // nzuimm{9:6] <= [10:7]
					((inst >> 4) & 0x4) | // nzuimm[2] <= [6]
					((inst >> 2) & 0x8) // nzuimm[3] <= [5]
			if nzuimm != 0 {
				return (nzuimm << 20) | (2 << 15) | ((rd + 8) << 7) | 0x13
			}
		case 1:
			// C.FLD for 32, 64-bit fld rd+8, offset(rs1+8)
			rd := (inst >> 2) & 0x7  // [4:2]
			rs1 := (inst >> 7) & 0x7 // [9:7]
			offset :=
				((inst >> 7) & 0x38) | // offset[5:3] <= [12:10]
					((inst << 1) & 0xc0) // offset[7:6] <= [6:5]
			return (offset << 20) | ((rs1 + 8) << 15) | (3 << 12) | ((rd + 8) << 7) | 0x7
		case 2:
			// C.LW lw rd+8, offset(rs1+8)
			rs1 := (inst >> 7) & 0x7 // [9:7]
			rd := (inst >> 2) & 0x7  // [4:2]
			offset :=
				((inst >> 7) & 0x38) | // offset[5:3] <= [12:10]
					((inst >> 4) & 0x4) | // offset[2] <= [6]
					((inst << 1) & 0x40) // offset[6] <= [5]
			return (offset << 20) | ((rs1 + 8) << 15) | (2 << 12) | ((rd + 8) << 7) | 0x3
		case 3:
			// C.LD in 64-bit mode ld rd+8, offset(rs1+8)
			rs1 := (inst >> 7) & 0x7 // [9:7]
			rd := (inst >> 2) & 0x7  // [4:2]
			offset :=
				((inst >> 7) & 0x38) | // offset[5:3] <= [12:10]
					((inst << 1) & 0xc0) // offset[7:6] <= [6:5]
			return (offset << 20) | ((rs1 + 8) << 15) | (3 << 12) | ((rd + 8) << 7) | 0x3
		case 4:
		// reserved
		case 5:
			// C.FSD fsd rs2+8, offset(rs1+8)
			rs1 := (inst >> 7) & 0x7 // [9:7]
			rs2 := (inst >> 2) & 0x7 // [4:2]
			offset :=
				((inst >> 7) & 0x38) | // uimm[5:3] <= [12:10]
					((inst << 1) & 0xc0) // uimm[7:6] <= [6:5]
			imm11_5 := (offset >> 5) & 0x7f
			imm4_0 := offset & 0x1f
			return (imm11_5 << 25) | ((rs2 + 8) << 20) | ((rs1 + 8) << 15) | (3 << 12) | (imm4_0 << 7) | 0x27
		case 6:
			// C.SW sw rs2+8, offset(rs1+8)
			rs1 := (inst >> 7) & 0x7 // [9:7]
			rs2 := (inst >> 2) & 0x7 // [4:2]
			offset :=
				((inst >> 7) & 0x38) | // offset[5:3] <= [12:10]
					((inst << 1) & 0x40) | // offset[6] <= [5]
					((inst >> 4) & 0x4) // offset[2] <= [6]
			imm11_5 := (offset >> 5) & 0x7f
			imm4_0 := offset & 0x1f
			return (imm11_5 << 25) | ((rs2 + 8) << 20) | ((rs1 + 8) << 15) | (2 << 12) | (imm4_0 << 7) | 0x23
		case 7:
			// C.SD sd rs2+8, offset(rs1+8)
			rs1 := (inst >> 7) & 0x7 // [9:7]
			rs2 := (inst >> 2) & 0x7 // [4:2]
			offset :=
				((inst >> 7) & 0x38) | // uimm[5:3] <= [12:10]
					((inst << 1) & 0xc0) // uimm[7:6] <= [6:5]
			imm11_5 := (offset >> 5) & 0x7f
			imm4_0 := offset & 0x1f
			return (imm11_5 << 25) | ((rs2 + 8) << 20) | ((rs1 + 8) << 15) | (3 << 12) | (imm4_0 << 7) | 0x23
		}
	case 1:
		switch funct3 {
		case 0:
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
		case 6:
		case 7:
		}
	case 2:
	}

	return 0x0
}

func (cpu *CPU) tick() {
	pc := cpu.pc
	if excp := cpu.run(); excp != nil {
		cpu.handleExcp(excp, pc)
	}

	//cpu.clint.tick(cpu.rcsr(mip))
	//cpu.disk.tick()
	//cpu.uart.tick()
	//cpu.plic.tick()
	cpu.handleIntr(cpu.pc)
	cpu.clock++
	cpu.wcsr(cycle, cpu.clock*8)
}

func (cpu *CPU) run() *trap {
	if cpu.wfi {
		if (cpu.rcsr(mie) & cpu.rcsr(mip)) != 0 {
			cpu.wfi = false
		}
		return nil
	}

	w, excp := cpu.fetch()
	if excp != nil {
		return excp
	}

	pc := cpu.pc
	if w&0x3 == 0x3 {
		cpu.pc += 4
	} else {
		cpu.pc += 2 // compressed
		w = cpu.decompress(w & 0xffff)
	}

	return cpu.exec(w, pc)
}

func (cpu *CPU) exec(raw, pc uint64) *trap {
	switch {
	case raw&0xfe00707f == 0x00000033: //"add"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)+cpu.rxreg(rs2))

	case raw&0x0000707f == 0x00000013: //"addi"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		cpu.wxreg(rd, imm+cpu.rxreg(rs1))

	case raw&0x0000707f == 0x0000001b: //"addiw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)+imm))))

	case raw&0xfe00707f == 0x0000003b: //"addw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)+cpu.rxreg(rs2)))))

	case raw&0xf800707f == 0x0000302f: //"amoadd.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.write(addr, t+cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0x0000202f: //"amoadd.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.write(addr, t+cpu.rxreg(rs2), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0x6000302f: //"amoand.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.write(addr, t&cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0x6000202f: //"amoand.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.write(addr, uint64(int64(int32(t)&int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

		// 11111000000000000111000001111111

	case raw&0xf800707f == 0xa000302f: // amomax.d
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int64(t) < int64(t2) {
			cpu.write(addr, uint64(int64(t2)), doubleword)
		} else {
			cpu.write(addr, uint64(int64(t)), doubleword)
		}

		cpu.wxreg(rd, uint64(int64(t)))

	case raw&0xf800707f == 0xa000202f: // amomax.w
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int32(t) < int32(t2) {
			cpu.write(addr, uint64(int64(int32(t2))), word)
		} else {
			cpu.write(addr, uint64(int64(int32(t))), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0xe000302f: //"amomaxu.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if t < t2 {
			cpu.write(addr, t2, doubleword)
		} else {
			cpu.write(addr, t, doubleword)
		}
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0xe000202f: //"amomaxu.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if uint32(t) < uint32(t2) {
			cpu.write(addr, uint64(uint32(t2)), word)
		} else {
			cpu.write(addr, uint64(uint32(t)), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0xc000302f: // amominu.d
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if t < t2 {
			cpu.write(addr, t, doubleword)
		} else {
			cpu.write(addr, t2, doubleword)
		}

		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0xc000202f: // amominu.w
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if uint32(t) < uint32(t2) {
			cpu.write(addr, uint64(uint32(t)), word)
		} else {
			cpu.write(addr, uint64(uint32(t2)), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0x8000302f: // amomin.d
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int64(t) < int64(t2) {
			cpu.write(addr, uint64(int64(t)), doubleword)
		} else {
			cpu.write(addr, uint64(int64(t2)), doubleword)
		}

		cpu.wxreg(rd, uint64(int64(t)))

	case raw&0xf800707f == 0x8000202f: // amomin.w
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		t2 := cpu.rxreg(rs2)

		if int32(t) < int32(t2) {
			cpu.write(addr, uint64(int64(int32(t))), word)
		} else {
			cpu.write(addr, uint64(int64(int32(t2))), word)
		}

		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0x4000302f: //"amoor.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.write(addr, t|cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0x4000202f: //"amoor.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.write(addr, uint64(int64(int32(t)|int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0x0800302f: //"amoswap.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.write(addr, cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0x0800202f: //"amoswap.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.write(addr, cpu.rxreg(rs2), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xf800707f == 0x2000302f: // amoxor.d
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.write(addr, t^cpu.rxreg(rs2), doubleword)
		cpu.wxreg(rd, t)

	case raw&0xf800707f == 0x2000202f: // amoxor.w
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.write(addr, uint64(int64(int32(t)^int32(cpu.rxreg(rs2)))), word)
		cpu.wxreg(rd, uint64(int64(int32(t))))

	case raw&0xfe00707f == 0x00007033: //"and"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)&cpu.rxreg(rs2))

	case raw&0x0000707f == 0x00007013: //"andi"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)&imm)

	case raw&0x0000007f == 0x00000017: //"auipc"
		imm := uint64(int64(int32(uint32(bits(raw, 31, 12) << 12))))
		rd := bits(raw, 11, 7)
		cpu.wxreg(rd, pc+imm)

	case raw&0x0000707f == 0x00000063: //"beq"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if cpu.rxreg(rs1) == cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00005063: //"bge"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if int64(cpu.rxreg(rs1)) >= int64(cpu.rxreg(rs2)) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00007063: //"bgeu"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if cpu.rxreg(rs1) >= cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00004063: //"blt"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if int64(cpu.rxreg(rs1)) < int64(cpu.rxreg(rs2)) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00006063: //"bltu"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if cpu.rxreg(rs1) < cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00001063: //"bne"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseBImm(raw)
		if cpu.rxreg(rs1) != cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00003073: //"csrrc"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t & ^(cpu.rxreg(rs1))
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00007073: //"csrrci"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t & ^(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00002073: //"csrrs"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t | cpu.rxreg(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00006073: //"csrrsi"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t | rs1
		cpu.wcsr(imm, v) // RS1 is zimm
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00001073: //"csrrw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := cpu.rxreg(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00005073: //"csrrwi"
		rd, imm, csr := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		csr = csr & 0b111111111111
		cpu.wxreg(rd, cpu.rcsr(csr))
		cpu.wcsr(csr, imm)

	case raw&0xfe00707f == 0x02004033: //"div"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int64(cpu.rxreg(rs1))
		divisor := int64(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.rcsr(fcsr)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.wcsr(fcsr, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else if dividend == math.MinInt64 && divisor == -1 {
			cpu.wxreg(rd, uint64(dividend))
		} else {
			cpu.wxreg(rd, uint64(dividend/divisor))
		}

	case raw&0xfe00707f == 0x02005033: //"divu"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := cpu.rxreg(rs1)
		divisor := cpu.rxreg(rs2)
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.rcsr(fcsr)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.wcsr(fcsr, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else {
			cpu.wxreg(rd, dividend/divisor)
		}

	case raw&0xfe00707f == 0x0200503b: //"divuw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := uint32(cpu.rxreg(rs1))
		divisor := uint32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.rcsr(fcsr)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.wcsr(fcsr, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else {
			cpu.wxreg(rd, uint64(int64(int32(dividend/divisor))))
		}

	case raw&0xfe00707f == 0x0200403b: //"divw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := int32(cpu.rxreg(rs1))
		divisor := int32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0. Set a special flag
			v := cpu.rcsr(fcsr)
			v = setBit(v, 3) // DZ (Divided by Zero flag)
			cpu.wcsr(fcsr, v)
			cpu.wxreg(rd, 0xffff_ffff_ffff_ffff)
		} else if dividend == math.MinInt32 && divisor == -1 {
			cpu.wxreg(rd, uint64(int64(dividend)))
		} else {
			cpu.wxreg(rd, uint64(int64(dividend/divisor)))
		}

	case raw&0xffffffff == 0x00100073: //"ebreak"
		return &trap{code: breakpoint}

	case raw&0xffffffff == 0x00000073: //"ecall"
		switch cpu.mode {
		case user:
			return &trap{code: ecallFromU}
		case supervisor:
			return &trap{code: ecallFromS}
		case machine:
			return &trap{code: ecallFromM}
		default:
			return &trap{code: illegalInst, value: raw}
		}

	case raw&0xfe00007f == 0x02000053: //"fadd.d"

	case raw&0xfff0007f == 0xd2200053: //"fcvt.d.l"

	case raw&0xfff0007f == 0x42000053: //"fcvt.d.s"

	case raw&0xfff0007f == 0xd2000053: //"fcvt.d.w"

	case raw&0xfff0007f == 0xd2100053: //"fcvt.d.wu"

	case raw&0xfff0007f == 0x40100053: //"fcvt.s.d"

	case raw&0xfff0007f == 0xc2000053: //"fcvt.w.d"

	case raw&0xfe00007f == 0x1a000053: //"fdiv.d"

	case raw&0x0000707f == 0x0000000f: //"fence"
		// do nothing because rv currently does not apply any optimizations and no fence is needed.

	case raw&0x0000707f == 0x0000100f: //"fence.i"
		// do nothing because rv currently does not apply any optimizations and no fence is needed.

	case raw&0xfe00707f == 0xa2002053: //"feq.d"

	case raw&0x0000707f == 0x00003007: //"fld"

	case raw&0xfe00707f == 0xa2000053: //"fle.d"

	case raw&0xfe00707f == 0xa2001053: //"flt.d"

	case raw&0x0000707f == 0x00002007: //"flw"

	case raw&0x0600007f == 0x02000043: //"fmadd.d"

	case raw&0xfe00007f == 0x12000053: //"fmul.d"

	case raw&0xfff0707f == 0xf2000053: //"fmv.d.x"

	case raw&0xfff0707f == 0xe2000053: //"fmv.x.d"

	case raw&0xfff0707f == 0xe0000053: //"fmv.x.w"

	case raw&0xfff0707f == 0xf0000053: //"fmv.w.x"

	case raw&0x0600007f == 0x0200004b: //"fnmsub.d"

	case raw&0x0000707f == 0x00003027: //"fsd"

	case raw&0xfe00707f == 0x22000053: //"fsgnj.d"

	case raw&0xfe00707f == 0x22002053: //"fsgnjx.d"

	case raw&0xfe00007f == 0x0a000053: //"fsub.d"

	case raw&0x0000707f == 0x00002027: //"fsw"

	case raw&0x0000007f == 0x0000006f: //"jal"
		rd, imm := bits(raw, 11, 7), parseJImm(raw)
		tmp := pc + 4
		cpu.wxreg(rd, tmp)
		cpu.pc = pc + imm

	case raw&0x0000707f == 0x00000067: //"jalr"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		tmp := pc + 4
		target := (cpu.rxreg(rs1) + imm) & ^uint64(1)
		cpu.pc = target
		cpu.wxreg(rd, tmp)

	case raw&0x0000707f == 0x00000003: //"lb"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 8)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int8(v))))

	case raw&0x0000707f == 0x00004003: //"lbu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 8)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, v)

	case raw&0x0000707f == 0x00003003: //"ld"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		addr := cpu.rxreg(rs1) + imm
		r, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, r)

	case raw&0x0000707f == 0x00001003: //"lh"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, halfword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int16(v))))

	case raw&0x0000707f == 0x00005003: //"lhu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, halfword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, v)

	case raw&0xf9f0707f == 0x1000302f: //"lr.d"
		rd, rs1 := bits(raw, 11, 7), bits(raw, 19, 15)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(t))))
		cpu.reserve(addr)

	case raw&0xf9f0707f == 0x1000202f: //"lr.w"
		rd, rs1 := bits(raw, 11, 7), bits(raw, 19, 15)
		addr := cpu.rxreg(rs1)
		t, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(t))))
		cpu.reserve(addr)

	case raw&0x0000007f == 0x00000037: //"lui"
		imm := uint64(int64(int32(uint32(bits(raw, 31, 12) << 12))))
		rd := bits(raw, 11, 7)

		cpu.wxreg(rd, imm)

	case raw&0x0000707f == 0x00002003: //"lw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 32)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(v))))

	case raw&0x0000707f == 0x00006003: //"lwu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		addr := cpu.rxreg(rs1) + imm
		r, excp := cpu.read(addr, word)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, r)

	case raw&0xfe00707f == 0x02000033: //"mul"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))*int64(cpu.rxreg(rs2))))

	case raw&0xfe00707f == 0x02001033: //"mulh"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		v1 := cpu.rxreg(rs1)
		v2 := cpu.rxreg(rs2)
		// multiply as signed * signed
		bv1 := big.NewInt(int64(v1))
		bv2 := big.NewInt(int64(v2))
		bv1.Mul(bv1, bv2) // bv1 = bv1 * bv2
		bv1.Rsh(bv1, 64)  // bv1 = bv1 >> 64
		cpu.wxreg(rd, bv1.Uint64())

	case raw&0xfe00707f == 0x02003033: //"mulhu"
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

	case raw&0xfe00707f == 0x02002033: //"mulhsu"
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

	case raw&0xfe00707f == 0x0200003b: //"mulw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))*int32(cpu.rxreg(rs2)))))

	case raw&0xffffffff == 0x30200073: //"mret"
		// First, set CSRs[MEPC] to program counter.
		cpu.pc = cpu.rcsr(mepc)

		// Then, Modify MSTATUS.

		mst := cpu.rcsr(mstatus)

		// Set CPU mode according to MPP
		switch bits(mst, 12, 11) {
		case 0b00:
			cpu.mode = user
			mst = clearBit(mst, 17)
		case 0b01:
			cpu.mode = supervisor
			mst = clearBit(mst, 17)
		case 0b11:
			cpu.mode = machine
		default:
			// should not happen
			panic("invalid CSR MPP")
		}

		mpie := bit(mst, 7)

		// set MPIE to MIE
		if mpie == 0 {
			mst = clearBit(mst, 3)
		} else {
			mst = setBit(mst, 3)
		}

		// set 1 to MPIE
		mst = setBit(mst, 7)

		// set 0 to MPP
		mst = clearBit(mst, 12)
		mst = clearBit(mst, 11)

		// update MSTATUS
		cpu.wcsr(mstatus, mst)

	case raw&0xfe00707f == 0x00006033: //"or"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)|cpu.rxreg(rs2))

	case raw&0x0000707f == 0x00006013: //"ori"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)|imm)

	case raw&0xfe00707f == 0x02006033: //"rem"
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

	case raw&0xfe00707f == 0x02007033: //"remu"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := cpu.rxreg(rs1)
		divisor := cpu.rxreg(rs2)
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, dividend)
		} else {
			cpu.wxreg(rd, dividend%divisor)
		}

	case raw&0xfe00707f == 0x0200703b: //"remuw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		dividend := uint32(cpu.rxreg(rs1))
		divisor := uint32(cpu.rxreg(rs2))
		if divisor == 0 {
			// Division by 0.
			cpu.wxreg(rd, uint64(int64(int32(dividend))))
		} else {
			cpu.wxreg(rd, uint64(int64(int32(dividend%divisor))))
		}

	case raw&0xfe00707f == 0x0200603b: //"remw"
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

	case raw&0x0000707f == 0x00000023: //"sb"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.write(addr, cpu.rxreg(rs2), byt)

	case raw&0xf800707f == 0x1800302f: //"sc.d"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)

		if cpu.reserved(addr) {
			// SC succeeds.
			cpu.write(addr, cpu.rxreg(rs2), doubleword)
			cpu.wxreg(rd, 0)
		} else {
			// SC fails.
			cpu.wxreg(rd, 1)
		}

		cpu.cancel(addr)

	case raw&0xf800707f == 0x1800202f: //"sc.w"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		addr := cpu.rxreg(rs1)

		if cpu.reserved(addr) {
			// SC succeeds.
			cpu.cancel(addr)
			cpu.write(addr, cpu.rxreg(rs2), word)
			cpu.wxreg(rd, 0)
		} else {
			// SC fails.
			cpu.cancel(addr)
			cpu.wxreg(rd, 1)
		}

	case raw&0x0000707f == 0x00003023: //"sd"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.write(addr, cpu.rxreg(rs2), doubleword)

	case raw&0xfe007fff == 0x12000073: //"sfence.vma"
		// do nothing because rv currently does not apply any optimizations and no fence is needed.

	case raw&0x0000707f == 0x00001023: //"sh"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.write(addr, cpu.rxreg(rs2), halfword)

	case raw&0xfe00707f == 0x00001033: //"sll"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b11_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)<<shamt)

	case raw&0xfc00707f == 0x00001013: //"slli"
		rd, rs1, shamt := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 25, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)<<shamt)

	case raw&0xfe00707f == 0x0000101b: //"slliw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)<<shamt))))

	case raw&0xfe00707f == 0x0000103b: //"sllw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)<<shamt))))

	case raw&0xfe00707f == 0x00002033: //"slt"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		var v uint64 = 0
		if int64(cpu.rxreg(rs1)) < int64(cpu.rxreg(rs2)) {
			v = 1
		}
		cpu.wxreg(rd, v)

	case raw&0x0000707f == 0x00002013: //"slti"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.rxreg(rs1)) < int64(imm) {
			v = 1
		}
		cpu.wxreg(rd, v)

	case raw&0x0000707f == 0x00003013: //"sltiu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		var v uint64 = 0
		// must compare as two's complement
		if cpu.rxreg(rs1) < imm {
			v = 1
		}
		cpu.wxreg(rd, v)

	case raw&0xfe00707f == 0x00003033: //"sltu"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		var v uint64 = 0
		if cpu.rxreg(rs1) < cpu.rxreg(rs2) {
			v = 1
		}
		cpu.wxreg(rd, v)

	case raw&0xfe00707f == 0x40005033: //"sra"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shift := cpu.rxreg(rs2) & 0b111111
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))>>shift))

	case raw&0xfc00707f == 0x40005013: //"srai"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))>>shamt))

	case raw&0xfc00707f == 0x4000501b: //"sraiw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))>>shamt)))

	case raw&0xfe00707f == 0x4000503b: //"sraw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1))>>shamt)))

	case raw&0xffffffff == 0x10200073: //"sret"
		cpu.pc = cpu.rcsr(sepc)

		// Then, Modify SSTATUS.

		sst := cpu.rcsr(sstatus)

		// Set CPU mode according to SPP
		switch bit(sst, 8) {
		case 0b0:
			cpu.mode = user
		case 0b1:
			cpu.mode = supervisor

			// MPRV must be set 0 if the mode is not Machine.
			if cpu.mode == supervisor {
				mst := cpu.rcsr(mstatus)
				mst = clearBit(mst, 17)
				cpu.wcsr(mstatus, mst)
			}
		default:
			// should not happen
			panic("invalid CSR SPP")
		}

		spie := bit(sstatus, 5)

		// set SPIE to SIE
		if spie == 0 {
			sst = clearBit(sstatus, 1)
		} else {
			sst = setBit(sstatus, 1)
		}

		// set 1 to SPIE
		sst = setBit(sst, 5)

		// set 0 to SPP
		sst = clearBit(sst, 8)

		// update SSTATUS
		cpu.wcsr(sstatus, sst)

	case raw&0xfe00707f == 0x00005033: //"srl"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shift := cpu.rxreg(rs2) & 0b11_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)>>shift)

	case raw&0xfc00707f == 0x00005013: //"srli"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)>>shamt)

	case raw&0xfc00707f == 0x0000501b: //"srliw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(uint32(cpu.rxreg(rs1))>>shamt))))

	case raw&0xfe00707f == 0x0000503b: //"srlw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		shamt := cpu.rxreg(rs2) & 0b1_1111
		cpu.wxreg(rd, uint64(int64(int32(uint32(cpu.rxreg(rs1))>>shamt))))

	case raw&0xfe00707f == 0x40000033: //"sub"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)-cpu.rxreg(rs2))

	case raw&0xfe00707f == 0x4000003b: //"subw"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, uint64(int64(int32(cpu.rxreg(rs1)-cpu.rxreg(rs2)))))

	case raw&0x0000707f == 0x00002023: //"sw"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), parseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.write(addr, cpu.rxreg(rs2), word)

	case raw&0xffffffff == 0x00200073: //"uret"
		ust := cpu.rcsr(ustatus)
		upie := bit(ust, 4)

		// set UPIE to UIE
		if upie == 0 {
			ust = clearBit(ust, 0)
		} else {
			ust = setBit(ust, 0)
		}

		// set 1 to SPIE
		ust = setBit(ust, 4)

		// update USTATUS
		cpu.wcsr(ustatus, ust)

	case raw&0xffffffff == 0x10500073: //"wfi"
		cpu.wfi = true

	case raw&0xfe00707f == 0x00004033: //"xor"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)^cpu.rxreg(rs2))

	case raw&0x0000707f == 0x00004013: //"xori"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), parseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)^imm)
	}

	return nil
}

func (cpu *CPU) getCause(code int, intr bool) uint64 {
	if !intr {
		return uint64(code)
	}

	intrbit := uint64(0x8000000000000000)
	if cpu.xlen == xlen32 {
		intrbit = 0x80000000
	}

	return intrbit + uint64(code)
}

func (cpu *CPU) handleTrap(trp *trap, curPC uint64, intr bool) bool {
	curMode := cpu.mode
	cause := cpu.getCause(trp.code, intr)

	var mdeleg, sdeleg uint64
	if intr {
		mdeleg = cpu.rcsr(mideleg)
		sdeleg = cpu.rcsr(sideleg)
	} else {
		mdeleg = cpu.rcsr(medeleg)
		sdeleg = cpu.rcsr(sedeleg)
	}

	pos := cause & 0xffff

	newMode := user
	if ((mdeleg >> pos) & 1) == 0 {
		newMode = machine
	} else if ((sdeleg >> pos) & 1) == 0 {
		newMode = supervisor
	}

	var curStatus uint64
	switch cpu.mode {
	case machine:
		curStatus = cpu.rcsr(mstatus)
	case supervisor:
		curStatus = cpu.rcsr(sstatus)
	case user:
		curStatus = cpu.rcsr(ustatus)
	}

	if intr {
		var ie uint64
		switch newMode {
		case machine:
			ie = cpu.rcsr(mie)
		case supervisor:
			ie = cpu.rcsr(sie)
		case user:
			ie = cpu.rcsr(uie)
		}

		curMIE := (curStatus >> 3) & 1
		curSIE := (curStatus >> 1) & 1
		curUIE := curStatus & 1

		msie := (ie >> 3) & 1
		ssie := (ie >> 1) & 1
		usie := ie & 1

		mtie := (ie >> 7) & 1
		stie := (ie >> 5) & 1
		utie := (ie >> 4) & 1

		meie := (ie >> 11) & 1
		seie := (ie >> 9) & 1
		ueie := (ie >> 8) & 1

		if newMode < curMode {
			return false
		} else if newMode == curMode {
			switch cpu.mode {
			case machine:
				if curMIE == 0 {
					return false
				}
			case supervisor:
				if curSIE == 0 {
					return false
				}
			case user:
				if curUIE == 0 {
					return false
				}
			}
		}

		switch trp.code {
		case userSoftwareIntr:
			if usie == 0 {
				return false
			}
		case supervisorSoftwareIntr:
			if ssie == 0 {
				return false
			}
		case machineSoftwareIntr:
			if msie == 0 {
				return false
			}
		case userTimerIntr:
			if utie == 0 {
				return false
			}
		case supervisorTimerIntr:
			if stie == 0 {
				return false
			}
		case machineTimerIntr:
			if mtie == 0 {
				return false
			}
		case userExternalIntr:
			if ueie == 0 {
				return false
			}
		case supervisorExternalIntr:
			if seie == 0 {
				return false
			}
		case machineExternalIntr:
			if meie == 0 {
				return false
			}
		}
	}

	cpu.mode = newMode

	var epcAddr, causeAddr, tvalAddr, tvecAddr uint64

	switch cpu.mode {
	case machine:
		epcAddr, causeAddr, tvalAddr, tvecAddr = mepc, mcause, mtval, mtvec
	case supervisor:
		epcAddr, causeAddr, tvalAddr, tvecAddr = sepc, scause, stval, stvec
	case user:
		epcAddr, causeAddr, tvalAddr, tvecAddr = uepc, ucause, utval, utvec
	}

	cpu.wcsr(epcAddr, curPC)
	cpu.wcsr(causeAddr, cause)
	cpu.wcsr(tvalAddr, trp.value) // might be 0 which is okay
	cpu.pc = cpu.rcsr(tvecAddr)

	if cpu.pc&0x3 != 0 {
		cpu.pc = (cpu.pc & ^uint64(0x3)) + 4*(cause&0xffff)
	}

	switch cpu.mode {
	case machine:
		status := cpu.rcsr(mstatus)
		mie := (status >> 3) & 1
		newStatus := (status & ^uint64(0x1888)) | (mie << uint64(7)) | (uint64(curMode) << 11)
		cpu.wcsr(mstatus, newStatus)
	case supervisor:
		status := cpu.rcsr(sstatus)
		sie := (status >> 1) & 1
		newStatus := (status & ^uint64(0x122)) | (sie << uint64(5)) | ((uint64(curMode) & 1) << 8)
		cpu.wcsr(sstatus, newStatus)
	case user:
		panic("unimplemented")
	}

	return true
}

func (cpu *CPU) handleExcp(trp *trap, curPC uint64) {
	cpu.handleTrap(trp, curPC, false)
}

func (cpu *CPU) handleIntr(pc uint64) {
	mint := cpu.rcsr(mip) & cpu.rcsr(mie)

	if mint&0x800 != 0 { // meip
		if cpu.handleTrap(&trap{code: machineExternalIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x7ff)
			cpu.wfi = false
			return
		}
	}

	if mint&0x008 != 0 { // msip
		if cpu.handleTrap(&trap{code: machineSoftwareIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x0111)
			cpu.wfi = false
			return
		}
	}

	if mint&0x080 != 0 { // mtip
		if cpu.handleTrap(&trap{code: machineTimerIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x7f)
			cpu.wfi = false
			return
		}
	}

	if mint&0x200 != 0 { // seip
		if cpu.handleTrap(&trap{code: supervisorExternalIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x1ff)
			cpu.wfi = false
			return
		}
	}

	if mint&0x002 != 0 { // ssip
		if cpu.handleTrap(&trap{code: supervisorSoftwareIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x1)
			cpu.wfi = false
			return
		}
	}

	if mint&0x020 != 0 { // stip
		if cpu.handleTrap(&trap{code: supervisorTimerIntr}, pc, true) {
			cpu.wcsr(mip, cpu.rcsr(mip)&0x1f)
			cpu.wfi = false
			return
		}
	}
}

func parseIImm(inst uint64) uint64 {
	// inst[31:20] -> immediate[11:0].
	return signExtend(bits(inst, 31, 20), 12)
}

func parseSImm(inst uint64) uint64 {
	// inst[31:25] -> immediate[11:5], inst[11:7] -> immediate[4:0].
	return signExtend((bits(inst, 11, 7) | bits(inst, 31, 25)<<5), 12)
}

func parseBImm(inst uint64) uint64 {
	// inst[31:25] -> immediate[12|10:5], inst[11:7] -> immediate[4:1|11].
	return signExtend((bit(inst, 31)<<12)|(bits(inst, 30, 25)<<5)|(bits(inst, 11, 8)<<1)|(bit(inst, 7)<<11), 13)
}

func parseJImm(inst uint64) uint64 {
	// inst[31:12] -> immediate[20|10:1|11|19:12].
	return signExtend((bit(inst, 31)<<20)|(bits(inst, 30, 21)<<1)|(bit(inst, 20)<<11)|(bits(inst, 19, 12)<<12), 21)
}

// returns val[i].
func bit(val uint64, pos int) uint64 {
	return bits(val, pos, pos)
}

// returns val[hi:lo].
func bits(val uint64, hi, lo int) uint64 {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

func setBit(val uint64, pos int) uint64 {
	return val | (1 << pos)
}

func clearBit(val uint64, pos int) uint64 {
	return val & ^(1 << pos)
}

func signExtend(v uint64, size int) uint64 {
	tmp := 64 - size
	return uint64((int64(v) << tmp) >> tmp)
}
