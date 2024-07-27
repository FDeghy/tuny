package core

import (
	"Tuny/portal/tcp"
	"Tuny/tools"
	"Tuny/transport"
	"net/netip"
	"time"

	"github.com/lesismal/nbio"
	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
)

func StartForwarder(c Config) error {
	logger = log.With().
		Str("loc", "Tunnel Forwarder").
		Logger()

	if c.LocalAddr == "" {
		c.LocalAddr = "0.0.0.0:0"
	}
	if c.TunnelAddr == "" {
		logger.Fatal().
			Msg("tunnel address is empty")
	}
	if c.DestAddr == "" {
		logger.Fatal().
			Msg("destination address is empty")
	}

	engine, err := tcp.StartTcpConnector(c.DestAddr, fonNewData, fonClose)
	if err != nil {
		return err
	}
	err = transport.StartKharej(c.TunnelAddr, c.LocalAddr, quicConf, handleStream(c.DestAddr, engine))
	if err != nil {
		return err
	}

	return nil
}

func handleStream(dstAddr string, engine *nbio.Engine) func(transport.ConnType, quic.Stream) {
	return func(t transport.ConnType, stream quic.Stream) {
		b := tools.GetBuffer()
		defer tools.PutBuffer(b)
		payload := b[:6]

		n, _ := stream.Read(payload)
		addrPort := &netip.AddrPort{}
		addrPort.UnmarshalBinary(payload[:n])

		// connect to dest
		if t == transport.IRAN_TO_KHAREJ_CONN {
			c, err := nbio.Dial("tcp", dstAddr) // only tcp for now
			if err != nil {
				log.Error().
					Err(err).
					Msg("nbio dial error")
				stream.Close()
				closeStrConn(c, addrPort.String())
				return
			}

			engine.AddConn(c)

			mutexDial.Lock()
			sess, ok := dialTable[addrPort.String()]
			if !ok {
				sess = &tunnel{}
			}
			sess.conn = c
			sess.uStream = stream
			c.SetSession(addrPort.String())
			dialTable[addrPort.String()] = sess
			mutexDial.Unlock()

			buff := tools.GetBuffer()
			defer tools.PutBuffer(buff)
			for {
				n, err := stream.Read(buff)
				if err != nil {
					logger.Debug().
						Err(err).
						Int64("stream id", int64(stream.StreamID())).
						Msg("ITK stream read error")
					closeStrConn(c, addrPort.String())
					return
				}

				_, err = c.Write(buff[:n])
				if err != nil {
					logger.Debug().
						Err(err).
						Int64("stream id", int64(stream.StreamID())).
						Msg("ITK nbio write error")
					closeStrConn(c, addrPort.String())
					return
				}

				stream.SetReadDeadline(time.Now().Add(MaxIdleConnection))
			}
		} else if t == transport.KHAREJ_TO_IRAN_CONN {
			mutexDial.Lock()
			sess, ok := dialTable[addrPort.String()]
			if !ok {
				sess = &tunnel{}
			}
			sess.dStream = stream
			dialTable[addrPort.String()] = sess
			mutexDial.Unlock()
		}
	}
}

func closeStrConn(c *nbio.Conn, key string) {
	mutexDial.Lock()
	defer mutexDial.Unlock()
	sess, ok := dialTable[key]
	if !ok {
		sess = &tunnel{}
	}
	if isClosed, _ := c.IsClosed(); c != nil && !isClosed {
		c.Close()
	}
	if sess.uStream != nil {
		sess.uStream.Close()
	}
	if sess.dStream != nil {
		sess.dStream.Close()
	}
	delete(dialTable, key)
}

func fonNewData(c *nbio.Conn, data []byte) {
	var sess *tunnel
	key, _ := c.Session().(string)

	i := 0

	for ; i < 10; i++ {
		mutexDial.RLock()
		s, ok := dialTable[key]
		mutexDial.RUnlock()
		if !ok {
			logger.Info().
				Msg("dialTable sess not found")
			c.Close()
			closeStrConn(c, key)
			return
		}
		if s.dStream != nil {
			sess = s
			break
		}
		logger.Debug().
			Msg("wait 1s for dStream")
		time.Sleep(time.Second)
	}

	if i == 10 {
		logger.Debug().
			Msg("dStream not found")
		closeStrConn(c, key)
		return
	}

	_, err := sess.dStream.Write(data)

	if err != nil {
		logger.Warn().
			Err(err).
			Msg("error write to dStream")
		closeStrConn(c, key)
		return
	}

	c.SetReadDeadline(time.Now().Add(MaxIdleConnection))
}

func fonClose(c *nbio.Conn, err error) {
	key, _ := c.Session().(string)
	logger.Info().
		Err(err).
		Str("key", key).
		Msg("close forwarder")
	closeStrConn(c, key)
}
