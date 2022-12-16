package main

type VirtIODisk struct{}

func NewVirtIODisk() *VirtIODisk {
	return nil
}

func (v *VirtIODisk) read(addr uint64) uint8 {
	return 0
}

func (v *VirtIODisk) write(addr uint64, value uint8) {
}
