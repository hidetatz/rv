package main

import (
	"fmt"
	"strings"
)

type Registers struct {
	// every register size is 64bit
	Regs [32]uint64
}

func NewRegisters() *Registers {
	return &Registers{Regs: [32]uint64{}}
}

func (r *Registers) Read(i uint64) uint64 {
	// assuming i is valid
	return r.Regs[i]
}

func (r *Registers) Write(i uint64, val uint64) {
	// x0 is always zero, the write should be discarded in that case
	if i != 0 {
		r.Regs[i] = val
	}
}

func (r *Registers) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, r := range r.Regs {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("x%d: %x", i, r))
	}
	sb.WriteString("]")
	return sb.String()
}

type FRegisters struct {
	Regs [32]float64
}

func NewFRegisters() *FRegisters {
	return &FRegisters{Regs: [32]float64{}}
}

func (fr *FRegisters) Read(i uint64) float64 {
	// assuming i is valid
	return fr.Regs[i]
}

func (fr *FRegisters) Write(i uint64, val float64) {
	// f0 is always zero, the write should be discarded in that case
	if i != 0 {
		fr.Regs[i] = val
	}
}
