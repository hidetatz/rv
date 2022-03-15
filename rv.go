package main

type RV struct {
	cpu *CPU
}

func New(prog []byte) (*RV, error) {
	cpu := NewCPU()

	pc, err := LoadElf(prog, cpu.Bus.Memory)
	if err != nil {
		return nil, err
	}

	cpu.PC = pc

	return &RV{cpu: cpu}, nil
}

func (r *RV) Start() {
	for {
		r.cpu.Run()
	}
}

func (r *RV) StartForTest(count int) {
	for i := 0; i < count; i++ {
		r.cpu.Run()
	}
}
