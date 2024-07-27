package transport

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
	"github.com/xtls/xray-core/common/protocol/tls/cert"
	xtls "github.com/xtls/xray-core/transport/internet/tls"
)

// IRAN

var (
	iranConnections = make([]qConnection, 0, 64)
	mutex           = &sync.RWMutex{}
	iranNewConn     = make(chan qConnection, 4)
)

func StartIran(localTunnelAddr string, quicConf *quic.Config) error {
	quicConfig = quicConf

	logger = log.With().
		Str("loc", "Iran Quic").
		Logger()

	conn, err := NewQuicConn(localTunnelAddr)
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

	go acceptConnenction(ln)
	go handleConnections()

	// create manage stream
	_, err = newManageStream()
	if err != nil {
		return fmt.Errorf("newManageStream %w", err)
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

		iranNewConn <- qConnection{
			Type:       UNKNOWN,
			Conn:       conn,
			RemoteAddr: conn.RemoteAddr(),
		}
	}
}

func haveManageConn() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	for _, c := range iranConnections {
		if c.Type == CREATE_MANAGE {
			return true
		}
	}
	return false
}

func handleConnections() {
	var conn qConnection
	i := 0
	for {
		conn = <-iranNewConn

		if !haveManageConn() {
			conn.Type = CREATE_MANAGE
			mutex.Lock()
			iranConnections = append(iranConnections, conn)
			mutex.Unlock()
			continue
		}

		if i == 0 {
			conn.Type = IRAN_TO_KHAREJ_CONN
			mutex.Lock()
			iranConnections = append(iranConnections, conn)
			log.Debug().
				Int("len of iranConnections", len(iranConnections)).
				Msg("new iran to kharej connction")
			mutex.Unlock()
			i = 1
		} else if i == 1 {
			conn.Type = KHAREJ_TO_IRAN_CONN
			mutex.Lock()
			iranConnections = append(iranConnections, conn)
			log.Debug().
				Int("len of iranConnections", len(iranConnections)).
				Msg("new kharej to iran connction")
			mutex.Unlock()
			i = 0
		}
	}
}

func GetStream(stype ConnType) (quic.Stream, error) {
	_, err := waitActiveConn()
	if err != nil {
		return nil, fmt.Errorf("waitActiveConn %w", err)
	}

	mutex.RLock()
	mStream := manageStream
	mutex.RUnlock()
	if mStream == nil {
		mStream, err = newManageStream()
		if err != nil {
			return nil, fmt.Errorf("newManageStream %w", err)
		}
	}

	mutex.RLock()
	defer mutex.RUnlock()
	for i, c := range iranConnections {
		if i == int(math.Max(0, float64(len(iranConnections)-4))) { // allways have 2 idle connection
			requestNewConnection(mStream)
		}
		if c.Type != stype {
			continue
		}
		stream, err := c.Conn.OpenStream()
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("OpenStream error")
			continue
		}
		stream.Write([]byte{byte(stype)})
		logger.Info().
			Uint8("type", byte(stype)).
			Int64("stream id", int64(stream.StreamID())).
			Msg("stream created")
		return stream, nil
	}

	return nil, fmt.Errorf("no connection available")
}

func waitActiveConn() (qConnection, error) {
	for {
		mutex.RLock()
		lenIrConn := len(iranConnections)
		mutex.RUnlock()

		if lenIrConn > 0 {
			mutex.RLock()
			defer mutex.RUnlock()
			return iranConnections[0], nil
		}
		log.Debug().
			Msg("wait 3s to get connection")
		time.Sleep(3 * time.Second)
	}
	//return qConnection{}, fmt.Errorf("timeout")
}

func newManageStream() (quic.Stream, error) {
	var stream quic.Stream

	_, err := waitActiveConn()
	if err != nil {
		return nil, fmt.Errorf("waitActiveConn %w", err)
	}

	mutex.Lock()
	for i, c := range iranConnections {
		if c.Type != UNKNOWN && c.Type != CREATE_MANAGE {
			continue
		}
		stream, err = c.Conn.OpenStream()
		if err != nil {
			stream = nil
			logger.Warn().
				Err(err).
				Msg("newManageStream open error")
			logger.Info().
				Msg("wait 1s to get newManageStream")
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		_, err = stream.Write([]byte{byte(CREATE_MANAGE)})
		if err != nil {
			stream = nil
			logger.Warn().
				Err(err).
				Msg("newManageStream write error")
			logger.Info().
				Msg("wait 1s to get newManageStream\n")
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		if manageStream != nil {
			manageStream.Close()
		}
		manageStream = stream
		c.Type = CREATE_MANAGE
		iranConnections[i] = c
		break
	}
	mutex.Unlock()
	if stream == nil {
		return nil, fmt.Errorf("cannot find manager connection to create mStream")
	}
	return stream, nil
}

func requestNewConnection(stream quic.Stream) {
	for {
		_, err := stream.Write([]byte{byte(CREATE_BICONN)})
		if err != nil {
			newManageStream()
			continue
		}
		log.Debug().
			Msg("requestNewConnection sent")
		break
	}
}
