package transport

import (
	"sync"
)

type PortPool struct {
	pool  map[uint16]bool
	mutex *sync.RWMutex
}

func NewPortPool() *PortPool {
	p := make(map[uint16]bool)
	for i := 1; i < 65535; i++ {
		p[uint16(i)] = true
	}
	return &PortPool{
		pool:  p,
		mutex: &sync.RWMutex{},
	}
}

func (p *PortPool) GetFreePort() uint16 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for port, free := range p.pool {
		if free {
			p.pool[port] = false
			return port
		}
	}
	panic("no free port found")
}

func (p *PortPool) FreePort(port uint16) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pool[port] = true
}
