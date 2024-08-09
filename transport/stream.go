package transport

import (
	"github.com/quic-go/quic-go"
)

type Stream struct {
	Stream quic.Stream
	conn   *qConnection
	closed bool
}

func (s *Stream) Close() {
	err := s.Stream.Close()
	if err != nil || s.closed {
		return
	}
	if s.conn == nil { // kharej
		return
	}

	s.closed = true
	s.conn.DecStream(1)
	if s.conn.Type == IRAN_TO_KHAREJ_CONN {
		iranConnectionsUpload.Update(s.conn)
	} else if s.conn.Type == KHAREJ_TO_IRAN_CONN {
		iranConnectionsDownload.Update(s.conn)
	}

	// if s.conn.GetNumStream() == 0 && iranConnectionsDownload.Len() > 4 && iranConnectionsUpload.Len() > 4 {
	// 	s.conn.Conn.CloseWithError(0, "no stream")
	// 	if s.conn.Type == IRAN_TO_KHAREJ_CONN {
	// 		iranConnectionsUpload.Delete(s.conn.index)
	// 	} else if s.conn.Type == KHAREJ_TO_IRAN_CONN {
	// 		iranConnectionsDownload.Delete(s.conn.index)
	// 	}
	// }
}
