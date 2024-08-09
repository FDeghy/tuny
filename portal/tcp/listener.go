package tcp

import (
	"github.com/lesismal/nbio"
	"github.com/rs/zerolog/log"
)

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

	engine.OnData(onNewData)

	engine.OnClose(onClose)

	err := engine.Start()
	if err != nil {
		return engine, err
	}

	logger.Info().
		Str("address", engine.Addrs[0]).
		Msg("nbio.Engine Started")

	return engine, nil
}
