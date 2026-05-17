package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"sync"
)

type void struct{}

type responder struct {
	conn   *net.Conn
	mu     sync.Mutex
	buffer []any
	sendCh chan void
	err    error
}

func newResponder(conn *net.Conn) *responder {
	r := &responder{
		buffer: make([]any, 0),
		conn:   conn,
		sendCh: make(chan void, 1),
	}
	go r.flusher()
	return r
}

func (r *responder) reply(x any) {
	r.mu.Lock()
	r.buffer = append(r.buffer, x)
	r.mu.Unlock()
}

func (r *responder) throw(what error) {
	r.mu.Lock()
	r.err = what
	r.mu.Unlock()
}

func (r *responder) flush() {
	select {
	case r.sendCh <- void{}:
	default:
	}
}

func (r *responder) flusher() {
    header := make([]byte, 16)
    binary.BigEndian.PutUint32(header[0:4], MAGIC_NUMBER)

    for range r.sendCh {
        r.mu.Lock()

        localBuffer := r.buffer
        localErr := r.err
        
        r.buffer = nil
        r.err = nil
        r.mu.Unlock()

        var dataBuf []byte
        var isError uint32

        if localErr != nil {
            isError = 1
            dataBuf = []byte(localErr.Error())
        } else {
            isError = 0
            var buf bytes.Buffer
            enc := gob.NewEncoder(&buf)
            for _, v := range localBuffer {
                if err := enc.Encode(v); err != nil {
                    continue
                }
            }
            dataBuf = buf.Bytes()
        }

        binary.BigEndian.PutUint64(header[4:12], uint64(len(dataBuf)))
        binary.BigEndian.PutUint32(header[12:16], isError)

        packet := append(header, dataBuf...)
        if _, err := (*r.conn).Write(packet); err != nil {
            return 
        }
    }
}