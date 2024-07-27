package core

import (
	"io"
	"net"
	"time"
)

type Config struct {
	TunnelAddr string
	LocalAddr  string
	DestAddr   string
}

type stream interface {
	io.ReadWriteCloser
	SetReadDeadline(t time.Time) error
}

type tunnel struct {
	conn    net.Conn
	uStream stream
	dStream stream
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
