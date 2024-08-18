package transport

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
	"github.com/xtls/xray-core/common/protocol/tls/cert"
	xtls "github.com/xtls/xray-core/transport/internet/tls"
)

// IRAN

var (
	iranConnectionsUpload   = &Connections{}
	iranConnectionsDownload = &Connections{}
	iranNewConn             = make(chan *qConnection, 16)
)

func StartIran(localTunnelAddr string, proto int, quicConf *quic.Config) error {
	quicConfig = quicConf

	logger = log.With().
		Str("loc", "Iran Quic").
		Logger()

	conn, err := NewQuicConn(localTunnelAddr, proto)
	if err != nil {
		return fmt.Errorf("transport.NewQuicConn: %w", err)
	}

	tr := quic.Transport{
		ConnectionIDLength: 12,
		Conn:               conn,
	}
	tlsConfig := &xtls.Config{
		Certificate: []*xtls.Certificate{
			xtls.ParseCertificate(
				cert.MustGenerate(
					nil,
					cert.DNSNames(internalDomain),
					cert.CommonName(internalDomain),
				),
			),
		},
	}
	ln, err := tr.Listen(
		tlsConfig.GetTLSConfig(),
		//generateTLSConfig(),
		quicConfig,
	)
	if err != nil {
		return fmt.Errorf("tr.Listen: %w", err)
	}

	log.Debug().
		Str("address", conn.LocalAddr().String()).
		Str("network", conn.LocalAddr().Network()).
		Msg("start listening")

	go acceptConnenction(ln)
	go handleConnections()

	// create manage stream
	for {
		err = newManageStream()
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("waiting for connection")
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func acceptConnenction(ln *quic.Listener) {
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			logger.Error().
				Err(err).
				Msg("ln.Accept error")
			return
		}

		log.Info().
			Msgf("new quic conn accept: %v", conn.RemoteAddr())

		iranNewConn <- &qConnection{
			Type:       UNKNOWN,
			Conn:       conn,
			RemoteAddr: conn.RemoteAddr(),
			mutex:      &sync.RWMutex{},
		}
	}
}

func handleConnections() {
	var conn *qConnection

	i := 0
	for {
		conn = <-iranNewConn

		if i == 0 {
			conn.Type = IRAN_TO_KHAREJ_CONN
			heap.Push(iranConnectionsUpload, conn)
			log.Debug().
				Int("len of iranConnectionsUpload", iranConnectionsUpload.Len()).
				Msg("new iran to kharej connction")
			i = 1
		} else if i == 1 {
			conn.Type = KHAREJ_TO_IRAN_CONN
			heap.Push(iranConnectionsDownload, conn)
			log.Debug().
				Int("len of iranConnectionsDownload", iranConnectionsDownload.Len()).
				Msg("new kharej to iran connction")
			i = 0
		}
	}
}

func GetStream(stype ConnType) (*Stream, error) {
	var conn *qConnection

	if iranConnectionsDownload.Len() <= 4 || iranConnectionsUpload.Len() <= 4 {
		requestNewConnection()
	}

	for i := 0; i < 5; i++ {
		if stype == IRAN_TO_KHAREJ_CONN {
			conn = iranConnectionsUpload.Get(0)
		} else if stype == KHAREJ_TO_IRAN_CONN {
			conn = iranConnectionsDownload.Get(0)
		}

		if conn != nil {
			break
		} else {
			requestNewConnection()
		}
	}

	if conn == nil {
		return nil, fmt.Errorf("cannot get connection from pool")
	}

	if conn.GetNumStream() == int(quicConfig.MaxIncomingStreams/2) {
		requestNewConnection()
	}

	stream, err := conn.Conn.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("OpenStream error: %w", err)
	}

	_, err = stream.Write([]byte{byte(stype)})
	if err != nil {
		return nil, fmt.Errorf("write stream type error: %w", err)
	}

	conn.AddStream(1)
	if stype == IRAN_TO_KHAREJ_CONN {
		iranConnectionsUpload.Update(conn)
	} else if stype == KHAREJ_TO_IRAN_CONN {
		iranConnectionsDownload.Update(conn)
	}

	logger.Info().
		Uint8("type", byte(stype)).
		Int64("stream id", int64(stream.StreamID())).
		Msg("stream created")

	return &Stream{
		Stream: stream,
		conn:   conn,
	}, nil
}

func newManageStream() error {
	qc := iranConnectionsUpload.Get(0)
	if qc == nil {
		return fmt.Errorf("cannot get a connection")
	}

	stream, err := qc.Conn.OpenStream()
	if err != nil {
		return fmt.Errorf("OpenStream error: %w", err)
	}
	_, err = stream.Write([]byte{byte(CREATE_MANAGE)})
	if err != nil {
		return fmt.Errorf("stream Write error: %w", err)
	}

	if manageStream != nil {
		manageStream.Close()
	}
	manageStream = &Stream{
		Stream: stream,
		conn:   qc,
	}

	qc.AddStream(1)
	iranConnectionsUpload.Update(qc)

	return nil
}

func requestNewConnection() {
	for {
		_, err := manageStream.Stream.Write([]byte{byte(CREATE_BICONN)})
		if err != nil {
			newManageStream()
			continue
		}
		log.Debug().
			Msg("requestNewConnection sent")
		break
	}
}
