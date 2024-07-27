package core

import (
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog"
)

var (
	logger    zerolog.Logger
	dialTable = make(map[string]*tunnel)
	mutexDial = &sync.RWMutex{}
	quicConf  = &quic.Config{
		KeepAlivePeriod:       15 * time.Second,
		HandshakeIdleTimeout:  8 * time.Second,
		MaxIdleTimeout:        MaxIdleConnection,
		MaxIncomingStreams:    32,
		MaxIncomingUniStreams: -1,
	}
)

const (
	MaxIdleConnection = 120 * time.Second
)
