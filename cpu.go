package main

import (
	"math"
	"math/big"
)

const (
	// mode
	user       = 0
	supervisor = 1
	machine    = 3

	// xlen. 128bit isn't supported in rv.
	xlen32 = 1
	xlen64 = 2

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
	svnone = 0
	sv32   = 1
	sv39   = 2
	sv48   = 3
)

type CPU struct {
	clock          uint64
	xlen           int
	mode           int
	wfi            bool
	pc             uint64
	addressingMode int
	ppn            uint64

	csr   [4096]uint64
	xregs [32]uint64
	fregs [32]float64
	lrsc  map[uint64]struct{}

	ram *Memory
}

func NewCPU() *CPU {
	return &CPU{
		clock:          0,
		xlen:           xlen64,
		mode:           machine,
		pc:             0,
		addressingMode: svnone,
		ppn:            0,

		csr:   [4096]uint64{},
		xregs: [32]uint64{},
		fregs: [32]float64{},
		lrsc:  make(map[uint64]struct{}),

		ram: NewMemory(),
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

func (cpu *CPU) fetch() (uint64, *Exception) {
	vAddr := cpu.pc
	if (vAddr & 0xfff) <= 0x1000-4 {
		eAddr := cpu.getEffectiveAddr(vAddr)
		pa, excp := cpu.translate(eAddr, maInst, cpu.mode)
		if excp != nil {
			return 0, ExcpInstructionPageFault(vAddr)
		}

		v := cpu.ram.Read(pa, word)

		return v, nil
	}

	data := uint64(0)
	for i := uint64(0); i < 4; i++ {
		eAddr := cpu.getEffectiveAddr(vAddr + 1)
		pa, excp := cpu.translate(eAddr, maInst, cpu.mode)
		if excp != nil {
			return 0, ExcpInstructionPageFault(vAddr)
		}

		v := cpu.ram.Read(pa, byt)
		data |= v << (i * 8)
	}

	return data, nil
}

func (cpu *CPU) read(addr uint64, size int) (uint64, *Exception) {
	if size == byt {
		eAddr := cpu.getEffectiveAddr(addr)
		pAddr, excp := cpu.translate(eAddr, maLoad, cpu.mode)
		if excp != nil {
			return 0, ExcpLoadPageFault(addr)
		}

		return cpu.ram.Read(pAddr, byt), nil
	}

	if (addr & 0xfff) <= 0x1000-uint64(size/8) {
		eAddr := cpu.getEffectiveAddr(addr)
		pAddr, excp := cpu.translate(eAddr, maLoad, cpu.mode)
		if excp != nil {
			return 0, ExcpLoadPageFault(addr)
		}

		return cpu.ram.Read(pAddr, size), nil
	}

	data := uint64(0)
	for i := uint64(0); i < uint64(size/8); i++ {
		eAddr := cpu.getEffectiveAddr(addr + i)
		pa, excp := cpu.translate(eAddr, maLoad, cpu.mode)
		if excp != nil {
			return 0, ExcpLoadPageFault(addr)
		}

		v := cpu.ram.Read(pa, byt)
		data |= v << (i * 8)
	}

	return data, nil
}

func (cpu *CPU) write(addr, val uint64, size int) *Exception {
	// Cancel reserved memory to make SC fail when an write is called
	// between LR and SC.
	if cpu.reserved(addr) {
		cpu.cancel(addr)
	}

	if size == byt {
		eAddr := cpu.getEffectiveAddr(addr)
		pAddr, excp := cpu.translate(eAddr, maStore, cpu.mode)
		if excp != nil {
			return ExcpStoreAMOPageFault(addr)
		}

		cpu.ram.Write(pAddr, val, byt)
	}

	// multiple bytes

	if (addr & 0xfff) <= 0x1000-uint64(size/8) {
		eAddr := cpu.getEffectiveAddr(addr)
		pAddr, excp := cpu.translate(eAddr, maStore, cpu.mode)
		if excp != nil {
			return ExcpStoreAMOPageFault(addr)
		}

		cpu.ram.Write(pAddr, val, size)
	}

	for i := uint64(0); i < uint64(size/8); i++ {
		eAddr := cpu.getEffectiveAddr(addr + i)
		pa, excp := cpu.translate(eAddr, maStore, cpu.mode)
		if excp != nil {
			return ExcpStoreAMOPageFault(addr)
		}

		cpu.ram.Write(pa, (val>>(i*8))&0xff, byt)
	}

	return nil
}

func (cpu *CPU) translate(vAddr uint64, ma int, curMode int) (uint64, *Exception) {
	if cpu.addressingMode == svnone {
		return vAddr, nil
	}

	if curMode == machine {
		if ma == maInst {
			return vAddr, nil
		}

		if bit(cpu.rcsr(mstatus), 17) == 0 {
			return vAddr, nil
		}

		newPrivMode := bits(cpu.rcsr(mstatus), 10, 9)
		if newPrivMode == machine {
			return vAddr, nil
		}

		return cpu.translate(vAddr, ma, int(newPrivMode))
	}

	vpns := []uint64{bits(vAddr, 20, 12), bits(vAddr, 29, 21), bits(vAddr, 38, 30)}
	return cpu.traversePage(vAddr, 2, cpu.ppn, vpns, ma)
}

func (cpu *CPU) traversePage(vAddr uint64, level int, parentPPN uint64, vpns []uint64, ma int) (uint64, *Exception) {
	fault := func() *Exception {
		switch ma {
		case maInst:
			return ExcpInstructionPageFault(vAddr)
		case maLoad:
			return ExcpLoadPageFault(vAddr)
		case maStore:
			return ExcpStoreAMOPageFault(vAddr)
		}

		return nil // should not come here
	}

	pteAddr := parentPPN*4096 + vpns[level]*8
	pte := cpu.ram.Read(pteAddr, doubleword)
	ppn := (pte >> 10) & 0xfffffffffff
	v, r, w, x, a, d := bit(pte, 0), bit(pte, 1), bit(pte, 2), bit(pte, 3), bit(pte, 6), bit(pte, 7)

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

		cpu.ram.Write(pteAddr, newPTE, doubleword)
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

	ppns := []uint64{
		(pte >> 10) & 0x1ff,
		(pte >> 19) & 0x1ff,
		(pte >> 28) & 0x3ff_ffff,
	}

	offset := vAddr & 0xfff
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

	//cpu.ram.tick()
	//cpu.handleIntr(cpu.pc)
	cpu.clock++
	//cpu.wcsr(cycle, cpu.clock*8)
}

func (cpu *CPU) run() *Exception {
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

func (cpu *CPU) exec(raw, pc uint64) *Exception {
	switch {
	case raw&0xfe00707f == 0x00000033: //"add"
		rd, rs1, rs2 := bits(raw, 11, 7), bits(raw, 19, 15), bits(raw, 24, 20)
		cpu.wxreg(rd, cpu.rxreg(rs1)+cpu.rxreg(rs2))

	case raw&0x0000707f == 0x00000013: //"addi"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, imm+cpu.rxreg(rs1))

	case raw&0x0000707f == 0x0000001b: //"addiw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)&imm)

	case raw&0x0000007f == 0x00000017: //"auipc"
		imm := uint64(int64(int32(uint32(bits(raw, 31, 12) << 12))))
		rd := bits(raw, 11, 7)
		cpu.wxreg(rd, pc+imm)

	case raw&0x0000707f == 0x00000063: //"beq"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) == cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00005063: //"bge"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if int64(cpu.rxreg(rs1)) >= int64(cpu.rxreg(rs2)) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00007063: //"bgeu"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) >= cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00004063: //"blt"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if int64(cpu.rxreg(rs1)) < int64(cpu.rxreg(rs2)) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00006063: //"bltu"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) < cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00001063: //"bne"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseBImm(raw)
		if cpu.rxreg(rs1) != cpu.rxreg(rs2) {
			cpu.pc = pc + imm
		}

	case raw&0x0000707f == 0x00003073: //"csrrc"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t & ^(cpu.rxreg(rs1))
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00007073: //"csrrci"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t & ^(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00002073: //"csrrs"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t | cpu.rxreg(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00006073: //"csrrsi"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := t | rs1
		cpu.wcsr(imm, v) // RS1 is zimm
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00001073: //"csrrw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		imm = imm & 0b111111111111
		t := cpu.rcsr(imm)
		v := cpu.rxreg(rs1)
		cpu.wcsr(imm, v)
		cpu.wxreg(rd, t)

	case raw&0x0000707f == 0x00005073: //"csrrwi"
		rd, imm, csr := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		return ExcpBreakpoint(pc)

	case raw&0xffffffff == 0x00000073: //"ecall"
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
		rd, imm := bits(raw, 11, 7), ParseJImm(raw)
		tmp := pc + 4
		cpu.wxreg(rd, tmp)
		cpu.pc = pc + imm

	case raw&0x0000707f == 0x00000067: //"jalr"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		tmp := pc + 4
		target := (cpu.rxreg(rs1) + imm) & ^uint64(1)
		cpu.pc = target
		cpu.wxreg(rd, tmp)

	case raw&0x0000707f == 0x00000003: //"lb"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 8)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int8(v))))

	case raw&0x0000707f == 0x00004003: //"lbu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 8)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, v)

	case raw&0x0000707f == 0x00003003: //"ld"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		addr := cpu.rxreg(rs1) + imm
		r, excp := cpu.read(addr, doubleword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, r)

	case raw&0x0000707f == 0x00001003: //"lh"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, halfword)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int16(v))))

	case raw&0x0000707f == 0x00005003: //"lhu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		v, excp := cpu.read(cpu.rxreg(rs1)+imm, 32)
		if excp != nil {
			return excp
		}
		cpu.wxreg(rd, uint64(int64(int32(v))))

	case raw&0x0000707f == 0x00006003: //"lwu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
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
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
		addr := cpu.rxreg(rs1) + imm
		cpu.write(addr, cpu.rxreg(rs2), doubleword)

	case raw&0xfe007fff == 0x12000073: //"sfence.vma"
		// do nothing because rv currently does not apply any optimizations and no fence is needed.

	case raw&0x0000707f == 0x00001023: //"sh"
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		var v uint64 = 0
		// must compare as two's complement
		if int64(cpu.rxreg(rs1)) < int64(imm) {
			v = 1
		}
		cpu.wxreg(rd, v)

	case raw&0x0000707f == 0x00003013: //"sltiu"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, uint64(int64(cpu.rxreg(rs1))>>shamt))

	case raw&0xfc00707f == 0x4000501b: //"sraiw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		shamt := imm & 0b1_1111
		cpu.wxreg(rd, cpu.rxreg(rs1)>>shamt)

	case raw&0xfc00707f == 0x0000501b: //"srliw"
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
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
		rs1, rs2, imm := bits(raw, 19, 15), bits(raw, 24, 20), ParseSImm(raw)
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
		rd, rs1, imm := bits(raw, 11, 7), bits(raw, 19, 15), ParseIImm(raw)
		cpu.wxreg(rd, cpu.rxreg(rs1)^imm)
	}

	return nil
}

// HandleException catches the raised exception and manipulates CSR and program counter based on
// the exception and CPU privilege mode.
func (cpu *CPU) handleExcp(excp *Exception, pc uint64) Trap {
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
		cpu.pc = cpu.rcsr(mtvec)
		if (cpu.pc & 0b11) != 0 {
			// Add 4 * cause if MTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.pc = (cpu.pc & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
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
		cpu.pc = cpu.rcsr(stvec)
		if (cpu.pc & 0b11) != 0 {
			// Add 4 * cause if STVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.pc = (cpu.pc & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
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
		cpu.pc = cpu.rcsr(utvec)
		if (cpu.pc & 0b11) != 0 {
			// Add 4 * cause if UTVEC has vector type address.
			// copied from: https://github.com/takahirox/riscv-rust/blob/master/src/cpu.rs#L625
			cpu.pc = (cpu.pc & ^uint64(0b11)) + uint64((4 * (cause * 0xffff)))
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
