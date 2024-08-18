package transport

import (
	"Tuny/tools"
	"context"
	"fmt"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
	xtls "github.com/xtls/xray-core/transport/internet/tls"
)

// KHAREJ

var (
	remoteTunnelAddr string
	handleStream     func(ConnType, quic.Stream)
	proto            int
)

func StartKharej(rTunAddr, localAddr string, ippr int, quicConf *quic.Config, hStream func(ConnType, quic.Stream)) error {
	quicConfig = quicConf
	remoteTunnelAddr = rTunAddr
	handleStream = hStream
	proto = ippr

	// dial, err := createConnection(localAddr) // manage connection
	// if err != nil {
	// 	return fmt.Errorf("createConnection %w", err)
	// }
	// //defer dial.CloseWithError(0, "bye")

	// go acceptStream(dial)

	err := createTunnelConnection()
	if err != nil {
		return fmt.Errorf("createTunnelConnection %w", err)
	}

	return nil
}

func createConnection(localAddr string, proto int) (quic.Connection, error) {
	conn, err := NewQuicConn(localAddr, proto)
	if err != nil {
		return nil, fmt.Errorf("transport.NewQuicConn: %w", err)
	}

	tr := quic.Transport{
		ConnectionIDLength: 12,
		Conn:               conn,
	}
	//addr, _ := net.ResolveUDPAddr("udp", remoteTunnelAddr)
	addr, err := net.ResolveIPAddr(fmt.Sprintf("ip:%v", proto), remoteTunnelAddr)
	log.Printf("addr: %+v, err: %v", addr, err)
	tlsConfig := &xtls.Config{
		ServerName:    internalDomain,
		AllowInsecure: true,
	}
	dial, err := tr.Dial(
		context.Background(),
		addr,
		tlsConfig.GetTLSConfig(),
		quicConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("conn.LocalAddr: %v, tr.Dial error: %w", conn.LocalAddr(), err)
	}

	return dial, nil
}

func acceptStream(conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			logger.Error().
				Err(err).
				Msg("conn.AcceptStream error")
			return
		}

		go onNewStream(stream)
	}
}

func onNewStream(stream quic.Stream) {
	logger.Debug().
		Int64("stream id", int64(stream.StreamID())).
		Msg("stream accepted")

	b := tools.GetBuffer()
	defer tools.PutBuffer(b)
	buff := b[:1]

	_, err := stream.Read(buff)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("first stream.Read error")
		stream.Close()
		return
	}

	switch ConnType(buff[0]) {
	case CREATE_MANAGE:
		logger.Info().
			Int64("stream id", int64(stream.StreamID())).
			Msg("new manager stream")

		go handleManageStream(stream)

	case IRAN_TO_KHAREJ_CONN:
		logger.Info().
			Int64("stream id", int64(stream.StreamID())).
			Msg("new IRAN_TO_KHAREJ_CONN stream")

		go handleStream(IRAN_TO_KHAREJ_CONN, stream)

	case KHAREJ_TO_IRAN_CONN:
		logger.Info().
			Int64("stream id", int64(stream.StreamID())).
			Msg("new KHAREJ_TO_IRAN_CONN stream")

		go handleStream(KHAREJ_TO_IRAN_CONN, stream)

	default:
		logger.Warn().
			Bytes("data", buff).
			Msg("stream byte[0] is not valid")
	}
}

func handleManageStream(stream quic.Stream) {
	b := tools.GetBuffer()
	defer tools.PutBuffer(b)
	buff := b[:1]

	for {
		if _, err := stream.Read(buff); err != nil {
			stream.Close()
			return
		} else {
			switch ConnType(buff[0]) {
			case CREATE_BICONN:
				createTunnelConnection()
			default:
				logger.Warn().
					Bytes("data", buff).
					Msg("invalid manage request")
			}
		}
		//stream.SetReadDeadline(time.Now().Add(10 * time.Second))
	}
}

func createTunnelConnection() error {
	// dial two connection
	conn1, err := createConnection("0.0.0.0", proto)
	if err != nil {
		return fmt.Errorf("createTunnelConnection: %w", err)
	}
	logger.Debug().
		Str("local address", conn1.LocalAddr().String()).
		Msg("new conn1 created")

	go acceptStream(conn1)

	conn2, err := createConnection("0.0.0.0", proto)
	if err != nil {
		return fmt.Errorf("createTunnelConnection: %w", err)
	}
	logger.Debug().
		Str("local address", conn2.LocalAddr().String()).
		Msg("new conn2 created")

	go acceptStream(conn2)

	return nil
}
