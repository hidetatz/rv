package main

const (
	virtioIrq = 1
	uartIrq   = 10
)

type Plic struct {
	clock          uint64
	irq            uint32
	enabled        uint64
	threshold      uint32
	ips            [1024]uint8
	priorities     [1024]uint32
	needsUpdateIrq bool
	virtioIpCache  bool
}

func NewPlic() *Plic {
	return &Plic{
		clock:          0,
		irq:            0,
		enabled:        0,
		threshold:      0,
		ips:            [1024]uint8{},
		priorities:     [1024]uint32{},
		needsUpdateIrq: false,
		virtioIpCache:  false,
	}
}

func (p *Plic) tick(virtioIp bool, uartIp bool, mip *uint64) {
	p.clock++
	if p.virtioIpCache != virtioIp {
		if virtioIp {
			p.setIp(virtioIrq)
		}
		p.virtioIpCache = virtioIp
	}

	if uartIp {
		p.setIp(uartIrq)
	}

	if p.needsUpdateIrq {
		p.updateIrq(mip)
		p.needsUpdateIrq = false
	}
}

func (p *Plic) updateIrq(mip *uint64) {
	virtioIp := ((p.ips[virtioIrq>>3] >> (virtioIrq & 7)) & 1) == 1
	uartIp := ((p.ips[uartIrq>>3] >> (uartIrq & 7)) & 1) == 1

	virtioPriority := p.priorities[virtioIrq]
	uartPriority := p.priorities[uartIrq]

	virtioEnabled := ((p.enabled >> virtioIrq) & 1) == 1
	uartEnabled := ((p.enabled >> uartIrq) & 1) == 1

	ips := []bool{virtioIp, uartIp}
	enables := []bool{virtioEnabled, uartEnabled}
	priorities := []uint32{virtioPriority, uartPriority}
	irqs := []uint32{virtioIrq, uartIrq}

	var irq uint32 = 0
	var priority uint32 = 0

	for i := 0; i < 2; i++ {
		if ips[i] && enables[i] && priorities[i] > p.threshold && priorities[i] > priority {
			irq = irqs[i]
			priority = priorities[i]
		}
	}

	p.irq = irq
	if p.irq != 0 {
		*mip |= 0x200
	}
}

func (p *Plic) setIp(irq uint32) {
	index := irq >> 3
	p.ips[index] = p.ips[index] | (1 << irq)
	p.needsUpdateIrq = true
}

func (p *Plic) clearIp(irq uint32) {
	index := irq >> 3
	p.ips[index] = p.ips[index] & ^(1 << irq)
	p.needsUpdateIrq = true
}

func (p *Plic) read(addr uint64) uint8 {
	return 0
}

func (p *Plic) write(addr uint64, value uint8) {
}
