package core

import (
	"Tuny/transport"
	"net"
)

type Config struct {
	TunnelAddr string
	LocalAddr  string
	DestAddr   string
}

type tunnel struct {
	conn    net.Conn
	uStream *transport.Stream
	dStream *transport.Stream
}

func (t tunnel) close() {
	if t.conn != nil {
		t.conn.Close()
	}
	if t.uStream != nil {
		t.uStream.Close()
	}
	if t.dStream != nil {
		t.dStream.Close()
	}
}
