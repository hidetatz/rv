package rv

type Registers struct {
	// every register size is 64bit
	Regs [32]uint64
}

func NewRegisters() *Registers {
	return &Registers{Regs: [32]uint64{}}
}

func (r *Registers) Read(i uint8) uint64 {
	// assuming i is valid
	return r.Regs[i]
}

func (r *Registers) Write(i uint8, val uint64) {
	// x0 is always zero, the write should be discarded in that case
	if i != 0 {
		r.Regs[i] = val
	}
}
