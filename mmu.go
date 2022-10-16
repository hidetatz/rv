package main

type MemoryAccessType int

const (
	MemoryAccessTypeInstruction MemoryAccessType = iota + 1
	MemoryAccessTypeLoad
	MemoryAccessTypeStore
)

type AddressingMode int

const (
	AddressingModeNone AddressingMode = iota
	AddressingModeSV32
	AddressingModeSV39
)

// MMU emulates memory management unit in a processor.
type MMU struct {
	ram *Memory
}

func NewMMU(xlen int) *MMU {
	return &MMU{
		ram: NewMemory(),
	}
}

func (mmu *MMU) Read(addr uint64, size int) uint64 {
	return mmu.ram.Read(addr, size)
}

func (mmu *MMU) Write(addr, val uint64, size int) {
	mmu.ram.Write(addr, val, size)
}
