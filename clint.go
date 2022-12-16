package main

type Clint struct{}

func NewClint() *Clint {
	return nil
}

func (c *Clint) read(addr uint64) uint8 {
	return 0
}

func (c *Clint) write(addr uint64, value uint8) {
}
