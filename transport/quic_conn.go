package transport

import (
	"Tuny/tools"
	"crypto/rand"
	"fmt"
	"math"
	mRand "math/rand"
	"net"
	"syscall"
	"time"
)

const (
	MIN_OBFS_BYTES = 16
	MAX_OBFS_BYTES = 32
)

type quicConn struct {
	conn *net.UDPConn
}

func NewQuicConn(IpPort string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", IpPort)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	return conn, err
	// return &quicConn{
	// 	conn: conn,
	// }, err
}

func (c *quicConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := c.conn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}
	data := b[:n]

	kLen := data[0]
	key := data[1 : 1+kLen]
	data = data[1+kLen:]
	for i := range key {
		data[i] = data[i] ^ key[i]
	}

	n = copy(b, data)

	return n, addr, nil
}

func (c *quicConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	buff := tools.GetBuffer()
	defer tools.PutBuffer(buff)

	if len(p) > len(buff) {
		return 0, fmt.Errorf("p larger than buff")
	}

	maxKeyLength := int(math.Min(float64(MAX_OBFS_BYTES), float64(len(p))))
	kLen := mRand.Intn(maxKeyLength-MIN_OBFS_BYTES) + MIN_OBFS_BYTES

	buff[0] = byte(kLen)
	rand.Read(buff[1 : 1+kLen])

	for i := 0; i < kLen; i++ {
		buff[1+kLen+i] = p[i] ^ buff[1+i]
	}
	copy(buff[1+kLen+kLen:], p[kLen:])

	return c.conn.WriteTo(buff[:1+kLen+len(p)], addr)
}

func (c *quicConn) Close() error {
	return c.conn.Close()
}

func (c *quicConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *quicConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *quicConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *quicConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *quicConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *quicConn) SetReadBuffer(bytes int) error {
	return c.conn.SetReadBuffer(bytes)
}

func (c *quicConn) SetWriteBuffer(bytes int) error {
	return c.conn.SetWriteBuffer(bytes)
}

func (c *quicConn) SyscallConn() (syscall.RawConn, error) {
	return c.conn.SyscallConn()
}

// func (c *quicConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
// 	buff := getBuffer()
// 	defer putBuffer(buff)

// 	n, oobn, flags, addr, err = c.conn.ReadMsgUDP(buff, oob)
// 	data := readDecode(buff[:n])
// 	n = copy(b, data)
// 	return
// }

// func (c *quicConn) WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error) {
// 	buff := getBuffer()
// 	defer putBuffer(buff)

// 	n = writeEncode(buff, b)

// 	n, oobn, err = c.conn.WriteMsgUDP(buff[:n], oob, addr)
// 	return
// }

// func (c *quicConn) Read(p []byte) (n int, err error) {
// 	n, _, _, _, err = c.ReadMsgUDP(p, nil)

// 	return
// }

// func (c *quicConn) Write(p []byte) (n int, err error) {
// 	n, _, _, _, err = c.ReadMsgUDP(p, nil)
// 	return
// }
