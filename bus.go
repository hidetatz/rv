package main

// Bus is a devices which is connected to the CPU,
// such as memory and some memory-mapped IO devices.
type Bus struct {
	Memory *Memory
}

// NewBus returns an initialized bus.
func NewBus() *Bus {
	return &Bus{
		Memory: NewMemory(),
	}
}

func (bus *Bus) Read(addr uint64, size int) uint64 {
	return bus.Memory.Read(addr, size)
}

func (bus *Bus) Write(addr, val uint64, size int) {
	bus.Memory.Write(addr, val, size)
}
