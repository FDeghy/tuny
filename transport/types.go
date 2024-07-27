package transport

import (
	"net"

	"github.com/quic-go/quic-go"
)

type ConnType uint8

type qConnection struct {
	Type       ConnType
	Conn       quic.Connection
	RemoteAddr net.Addr
}
