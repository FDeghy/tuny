package transport

import (
	"container/heap"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

var connMutex = &sync.RWMutex{}

type ConnType uint8

type qConnection struct {
	index      int
	Type       ConnType
	Conn       quic.Connection
	RemoteAddr net.Addr
	NumStreams int
	mutex      *sync.RWMutex
}

type Connections []*qConnection

func (c Connections) Len() int {
	connMutex.RLock()
	defer connMutex.RUnlock()
	return len(c)
}

func (c Connections) Less(i, j int) bool {
	connMutex.RLock()
	defer connMutex.RUnlock()
	return c[i].NumStreams < c[j].NumStreams
}

func (c Connections) Swap(i, j int) {
	connMutex.Lock()
	defer connMutex.Unlock()
	c[i], c[j] = c[j], c[i]
	c[i].index = i
	c[j].index = j
}

func (c *Connections) Push(x any) {
	connMutex.Lock()
	defer connMutex.Unlock()
	n := len(*c)
	qc := x.(*qConnection)
	qc.index = n
	*c = append(*c, qc)
}

func (c *Connections) Pop() any {
	connMutex.Lock()
	defer connMutex.Unlock()
	old := *c
	n := len(old)
	qc := old[n-1]
	old[n-1] = nil // avoid memory leak
	qc.index = -1  // for safety
	*c = old[:n-1]
	return qc
}

func (c *Connections) Update(qc *qConnection) {
	heap.Fix(c, qc.index)
}

func (c Connections) Get(i int) *qConnection {
	if c.Len() == 0 {
		return nil
	}
	connMutex.RLock()
	defer connMutex.RUnlock()
	return c[i]
}

func (c *Connections) Delete(i int) {
	connMutex.Lock()
	defer connMutex.Unlock()
	heap.Remove(c, i)
}

func (qc *qConnection) AddStream(i int) {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()
	qc.NumStreams += i
}

func (qc *qConnection) DecStream(i int) {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()
	qc.NumStreams -= i
}

func (qc *qConnection) GetNumStream() int {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()
	return qc.NumStreams
}
