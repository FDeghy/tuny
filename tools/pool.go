package tools

import "sync"

const BufferSize = 4096

var (
	pool = &sync.Pool{
		New: func() any {
			return make([]byte, BufferSize)
		},
	}
)

func GetBuffer() []byte {
	return pool.Get().([]byte)
}

func PutBuffer(b []byte) {
	pool.Put(b[:BufferSize])
}
