package tcp

import (
	"github.com/lesismal/nbio"
	"github.com/rs/zerolog/log"
)

func StartTcpConnector(
	destAddr string,
	onNewData func(*nbio.Conn, []byte),
	onClose func(*nbio.Conn, error),
) (*nbio.Engine, error) {

	logger = log.With().
		Str("loc", "TCP Connector").
		Logger()

	engine := nbio.NewEngine(nbio.Config{
		Name: "TcpConnector",
	})

	engine.OnData(onNewData)

	engine.OnClose(onClose)

	err := engine.Start()
	if err != nil {
		return engine, err
	}

	logger.Info().
		Msg("nbio.Engine Started")

	return engine, nil
}
