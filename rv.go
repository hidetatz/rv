package rv

type RV struct {
	cpu  *CPU
	prog []uint8
}

func New(prog []byte) *RV {
	return &RV{cpu: NewCPU(), prog: prog}
}

func (r *RV) Start() {
	r.cpu.Bus.Memory.Set(r.prog)
	for {
		r.cpu.Run()
	}
}
