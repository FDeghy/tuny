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

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	MIN_OBFS_BYTES = 8
	MAX_OBFS_BYTES = 16
)

var (
	spp = NewPortPool()
)

type quicConn struct {
	conn  *net.IPConn
	sport uint16
}

func NewQuicConn(IpPort string, proto int) (*quicConn, error) {
	// addr, err := net.ResolveUDPAddr("udp", IpPort)
	// if err != nil {
	// 	return nil, err
	// }
	// conn, err := net.ListenUDP("udp", addr)

	addr, _ := net.ResolveIPAddr(fmt.Sprintf("ip:%v", proto), IpPort)
	conn, err := net.ListenIP(fmt.Sprintf("ip:%v", proto), addr)
	//return conn, err
	return &quicConn{
		conn:  conn,
		sport: spp.GetFreePort(),
	}, err
}

func (c *quicConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := c.conn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}
	data := b[:n]

	var payload []byte
	var ipv4 layers.IPv4
	err = ipv4.DecodeFromBytes(data, gopacket.NilDecodeFeedback)
	if err != nil {
		payload = data
	} else {
		payload = ipv4.Payload
	}

	kLen := payload[0]
	key := payload[1 : 1+kLen]
	payload = payload[1+kLen:]
	for i := range key {
		payload[i] = payload[i] ^ key[i]
	}

	n = copy(b, payload)

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
	spp.FreePort(c.sport)
	return c.conn.Close()
}

func (c *quicConn) LocalAddr() net.Addr {
	return c
}

func (c *quicConn) Network() string {
	return c.conn.LocalAddr().Network()
}

func (c *quicConn) String() string {
	return fmt.Sprintf("%v:%v", c.conn.LocalAddr(), c.sport)
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
