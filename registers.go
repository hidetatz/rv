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

var XRegNames = map[string]int{
	"zero": 0,
	"ra":   1,
	"sp":   2,
	"gp":   3,
	"tp":   4,
	"t0":   5,
	"t1":   6,
	"t2":   7,
	"s0":   8,
	"s1":   9,
	"a0":   10,
	"a1":   11,
	"a2":   12,
	"a3":   13,
	"a4":   14,
	"a5":   15,
	"a6":   16,
	"a7":   17,
	"s2":   18,
	"s3":   19,
	"s4":   20,
	"s5":   21,
	"s6":   22,
	"s7":   23,
	"s8":   24,
	"s9":   25,
	"s10":  26,
	"s11":  27,
	"t3":   28,
	"t4":   29,
	"t5":   30,
	"t6":   31,
}

func XRegNameToIndex(name string) int {
	return XRegNames[name]
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
