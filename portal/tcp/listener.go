package tcp

import (
	"Tuny/tools"
	"fmt"
	"net"
	"time"

	"github.com/lesismal/nbio"
	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
)

type tcpServerConn struct {
	conn   *nbio.Conn
	stream net.Conn
}

func StartTcpServer(
	localAddr string,
	onNewConn func(*nbio.Conn),
	onNewData func(*nbio.Conn, []byte),
	onClose func(*nbio.Conn, error),
) (*nbio.Engine, error) {

	logger = log.With().
		Str("loc", "TCP Server").
		Logger()

	engine := nbio.NewEngine(nbio.Config{
		Name:               "TcpServer",
		Network:            nbio.NETWORK_TCP,
		Addrs:              []string{localAddr},
		MaxWriteBufferSize: MaxWriteBufferSize,
	})

	engine.OnOpen(onNewConn)
	// engine.OnOpen(func(c *nbio.Conn) {
	// 	stream, err := getStream()
	// 	if err != nil {
	// 		logger.Warn().
	// 			Err(err).
	// 			Msg("failed to getStream")
	// 		c.Close()
	// 		return
	// 	}

	// 	ts := &tcpServerConn{
	// 		conn:   c,
	// 		stream: stream,
	// 	}
	// 	c.SetSession(ts)

	// 	ts.handleNewConnection()
	// })

	engine.OnData(onNewData)
	// engine.OnData(func(c *nbio.Conn, data []byte) {
	// 	ts := c.Session().(*tcpServerConn)
	// 	ts.handleUpStream(data)
	// })

	engine.OnClose(onClose)
	// engine.OnClose(func(c *nbio.Conn, err error) {
	// 	ts := c.Session().(*tcpServerConn)
	// 	logger.Info().
	// 		Msg("OnClose connection")
	// 	ts.Close()
	// })

	err := engine.Start()
	if err != nil {
		return engine, err
	}

	logger.Info().
		Str("address", engine.Addrs[0]).
		Msg("nbio.Engine Started")

	return engine, nil
}

func (ts *tcpServerConn) handleNewConnection() {
	// add protocols in tunnel
	var id string
	switch stream := ts.stream.(type) {
	case quic.Stream:
		id = fmt.Sprintf("%v", stream.StreamID())
	default:
		id = stream.RemoteAddr().String()
	}

	logger.Info().
		Str("stream id", id).
		Msg("new connection and stream")

	go ts.handleDownStream()

	ts.conn.SetReadDeadline(time.Now().Add(MaxIdleConnection))
}

func (ts *tcpServerConn) handleDownStream() {
	buff := tools.GetBuffer()
	defer tools.PutBuffer(buff)

	n, err := ts.stream.Read(buff)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("Read DownStream from tunnel")
		ts.Close()
		return
	}

	_, err = ts.conn.Write(buff[:n])
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("Write DownConn to user")
		ts.Close()
		return
	}

	ts.stream.SetReadDeadline(time.Now().Add(MaxIdleConnection))
}

func (ts *tcpServerConn) handleUpStream(data []byte) {
	_, err := ts.stream.Write(data)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("Write UpStream to tunnel")
		ts.Close()
		return
	}

	ts.conn.SetReadDeadline(time.Now().Add(MaxIdleConnection))
}

func (ts *tcpServerConn) Close() {
	logger.Info().
		Msg("closing connection and stream")
	ts.conn.Close()
	ts.stream.Close()
}
