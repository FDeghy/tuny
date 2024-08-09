package core

import (
	"Tuny/portal/tcp"
	"Tuny/tools"
	"Tuny/transport"
	"net/netip"
	"time"

	"github.com/lesismal/nbio"
	"github.com/rs/zerolog/log"
)

func StartListener(c Config) error {
	logger = log.With().
		Str("loc", "Tunnel Listener").
		Logger()

	if c.LocalAddr == "" {
		c.LocalAddr = "0.0.0.0:0"
	}
	if c.TunnelAddr == "" {
		c.TunnelAddr = "0.0.0.0:4433"
	}

	err := transport.StartIran(c.TunnelAddr, quicConf)
	if err != nil {
		return err
	}
	_, err = tcp.StartTcpServer(c.LocalAddr, lonNewConn, lonNewData, lonClose)
	if err != nil {
		return err
	}

	return nil
}

func lonNewConn(c *nbio.Conn) {
	// get uStream & dStream
	uStream, err := transport.GetStream(transport.IRAN_TO_KHAREJ_CONN)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("failed to getStream")
		c.Close()
		return
	}
	dStream, err := transport.GetStream(transport.KHAREJ_TO_IRAN_CONN)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("failed to getStream")
		uStream.Close()
		c.Close()
		return
	}
	t := &tunnel{
		conn:    c,
		uStream: uStream,
		dStream: dStream,
	}
	c.SetSession(t)

	addr, _ := netip.ParseAddrPort(c.RemoteAddr().String())
	addrBytes, _ := addr.MarshalBinary()
	dStream.Stream.Write(addrBytes)
	uStream.Stream.Write(addrBytes)

	dStream.Stream.SetReadDeadline(time.Now().Add(MaxIdleConnection))
	c.SetReadDeadline(time.Now().Add(MaxIdleConnection))
	go t.handleDownStream()
}

// user -> tunnel
func lonNewData(c *nbio.Conn, data []byte) {
	t := c.Session().(*tunnel)
	_, err := t.uStream.Stream.Write(data)
	if err != nil {
		logger.Info().
			Err(err).
			Msg("onNewData write uStream error")
		t.close()
		return
	}
	t.conn.SetReadDeadline(time.Now().Add(MaxIdleConnection))
}

// tunnel -> user
func (t *tunnel) handleDownStream() {
	buff := tools.GetBuffer()
	defer tools.PutBuffer(buff)
	for {
		n, err := t.dStream.Stream.Read(buff)
		if err != nil {
			logger.Info().
				Err(err).
				Msg("handleDownStream read dStream error")
			t.close()
			return
		}
		_, err = t.conn.Write(buff[:n])
		if err != nil {
			logger.Info().
				Err(err).
				Msg("handleDownStream write conn error")
			t.close()
			return
		}
		t.dStream.Stream.SetReadDeadline(time.Now().Add(MaxIdleConnection))
	}
}

func lonClose(c *nbio.Conn, err error) {
	t, ok := c.Session().(*tunnel)
	if ok {
		t.close()
	}
}
