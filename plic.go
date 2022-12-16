package main

type Plic struct{}

func NewPlic() *Plic {
	return nil
}

func (p *Plic) read(addr uint64) uint8 {
	return 0
}

func (p *Plic) write(addr uint64, value uint8) {
}
