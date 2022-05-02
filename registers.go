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
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%s:0x%x", XRegNames[i], r))
	}
	sb.WriteString("]")
	return sb.String()
}

var XRegIndices = map[string]int{
	"zero": 0, "ra": 1, "sp": 2, "gp": 3, "tp": 4,
	"t0": 5, "t1": 6, "t2": 7,
	"s0": 8, "s1": 9,
	"a0": 10, "a1": 11, "a2": 12, "a3": 13, "a4": 14, "a5": 15, "a6": 16, "a7": 17,
	"s2": 18, "s3": 19, "s4": 20, "s5": 21, "s6": 22, "s7": 23, "s8": 24, "s9": 25, "s10": 26, "s11": 27,
	"t3": 28, "t4": 29, "t5": 30, "t6": 31,
}

var XRegNames = map[int]string{
	0: "zero", 1: "ra", 2: "sp", 3: "gp", 4: "tp",
	5: "t0", 6: "t1", 7: "t2",
	8: "s0", 9: "s1",
	10: "a0", 11: "a1", 12: "a2", 13: "a3", 14: "a4", 15: "a5", 16: "a6", 17: "a7",
	18: "s2", 19: "s3", 20: "s4", 21: "s5", 22: "s6", 23: "s7", 24: "s8", 25: "s9", 26: "s10", 27: "s11",
	28: "t3", 29: "t4", 30: "t5", 31: "t6",
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

func (r *FRegisters) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, r := range r.Regs {
		if i != 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%.1f", r))
	}
	sb.WriteString("]")
	return sb.String()
}
