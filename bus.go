package main

type bus struct {
	ram *Memory
	// TODO: add devices
}

func (bus *bus) Read(addr uint64, size int) uint64 {
	return bus.ram.Read(addr, size)
}

func (bus *bus) Write(addr, val uint64, size int) {
	bus.ram.Write(addr, val, size)
}
