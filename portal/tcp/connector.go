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
	// engine.OnClose(func(c *nbio.Conn, err error) {
	// 	mutexDial.RLock()
	// 	key, _ := c.Session().(string)
	// 	sess, ok := dialTable[key]
	// 	mutexDial.RUnlock()
	// 	if !ok {
	// 		sess = &session{}
	// 	}
	// 	closeStrConn(sess, c, key)
	// })

	err := engine.Start()
	if err != nil {
		return engine, err
	}

	logger.Info().
		Msg("nbio.Engine Started")

	return engine, nil
}

// func onNewStream() func(c *nbio.Conn, data []byte) {
// 	return func(c *nbio.Conn, data []byte) {
// 		mutexDial.RLock()
// 		key, _ := c.Session().(string)
// 		sess, ok := dialTable[key]
// 		mutexDial.RUnlock()
// 		if !ok {
// 			log.Printf("dialTable sess not found\n")
// 			c.Close()
// 			closeStrConn(sess, c, key)
// 			return
// 		}

// 		i := 0
// 		for ; i < 10; i++ {
// 			mutexDial.RLock()
// 			sess = dialTable[key]
// 			mutexDial.RUnlock()
// 			if sess.dStream != nil {
// 				break
// 			}
// 			log.Printf("wait 1s for dStream\n")
// 			time.Sleep(time.Second)
// 		}
// 		if i == 10 {
// 			log.Printf("dStream not found\n")
// 			closeStrConn(sess, c, key)
// 			return
// 		}
// 		_, err := sess.dStream.Write(data)
// 		if err != nil {
// 			log.Printf("error write to dStream: %v\n", err)
// 			closeStrConn(sess, c, key)
// 			return
// 		}
// 	}
// }
