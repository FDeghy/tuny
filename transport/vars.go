package transport

import (
	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog"
)

var (
	internalDomain         = "seyed.dev"
	manageStream   *Stream = nil
	quicConfig     *quic.Config
	logger         zerolog.Logger
)

const (
	// byte[0]
	IRAN_TO_KHAREJ_CONN ConnType = iota // upload to kharej
	KHAREJ_TO_IRAN_CONN                 // download from kharej
	CREATE_MANAGE
	CREATE_BICONN
	FINISH
	UNKNOWN
)

// func isActive(s quic.Connection) bool {
// 	select {
// 	case <-s.Context().Done():
// 		return false
// 	default:
// 		return true
// 	}
// }

// func removeInactiveConnections() { // TODO: optimze this to also check the streams
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	irConns := make([]qConnection, 0, 64)
// 	for _, v := range iranConnections {
// 		if isActive(v.Conn) {
// 			irConns = append(irConns, v)
// 			continue
// 		}

// 		log.Printf("closing tunnel %v", v.RemoteAddr.String())
// 		v.Conn.CloseWithError(0, "")
// 	}
// 	iranConnections = irConns

// 	for i := range connections {
// 		if !isActive(i.Conn) {
// 			i.Conn.CloseWithError(0, "")
// 			log.Printf("closing tunnel %v", i.RemoteAddr.String())
// 			delete(connections, i)
// 		}
// 	}
// }
