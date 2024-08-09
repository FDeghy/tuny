package transport

import "github.com/quic-go/quic-go"

type Stream struct {
	Stream quic.Stream
	conn   *qConnection
}

func (s Stream) Close() {
	s.Stream.Close()
	if s.conn == nil { // kharej
		return
	}
	s.conn.DecStream(1)
	if s.conn.Type == IRAN_TO_KHAREJ_CONN {
		iranConnectionsUpload.Update(s.conn)
	} else if s.conn.Type == KHAREJ_TO_IRAN_CONN {
		iranConnectionsDownload.Update(s.conn)
	}
	if s.conn.GetNumStream() == 0 {
		s.conn.Conn.CloseWithError(0, "no stream")
		if s.conn.Type == IRAN_TO_KHAREJ_CONN {
			iranConnectionsUpload.Delete(s.conn.index)
		} else if s.conn.Type == KHAREJ_TO_IRAN_CONN {
			iranConnectionsDownload.Delete(s.conn.index)
		}
	}
}
