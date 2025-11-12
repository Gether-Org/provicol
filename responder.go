package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"sync"
)

type void struct{}

type Responder struct {
	conn   *net.Conn
	mu     sync.Mutex
	buffer []any
	sendCh chan void
}

func newResponder(conn *net.Conn) (*Responder) {
	r := &Responder{
		buffer: make([]any, 0),
		conn:   conn,
		sendCh: make(chan void, 1),
	}
	go r.flusher()
	return r
}

func (r *Responder) Reply(x any) {
	r.mu.Lock()
	r.buffer = append(r.buffer, x)
	r.mu.Unlock()
}

func (r *Responder) Flush() {
	select {
	case r.sendCh <- void{}:
	default:
	}
}

func (r *Responder) flusher() {
	for range r.sendCh {
		r.mu.Lock()
		if len(r.buffer) == 0 {
			r.mu.Unlock()
			continue
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		for _, v := range r.buffer {
			if err := enc.Encode(v); err != nil {
				continue
			}
		}

		size := uint32(buf.Len())
		sizeBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBuf, size)

		(*r.conn).Write(append(sizeBuf, buf.Bytes()...))
		r.buffer = nil
		r.mu.Unlock()
	}
}
