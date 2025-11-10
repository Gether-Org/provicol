package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"sync"
)

type Responder struct {
    conn   net.Conn
    mu     sync.Mutex
    buffer []any           // accumulate responses
    sendCh chan struct{}   // signal pour envoyer le paquet
}

func NewResponder(conn net.Conn) *Responder {
    r := &Responder{
        conn:   conn,
        buffer: make([]any, 0),
        sendCh: make(chan struct{}, 1),
    }
    go r.flusher() // goroutine qui écoute les flush
    return r
}

// Reply n'envoie pas directement, juste stocke
func (r *Responder) Reply(x any) {
    r.mu.Lock()
    r.buffer = append(r.buffer, x)
    r.mu.Unlock()
}

// Flush force l'envoi
func (r *Responder) Flush() {
    select {
    case r.sendCh <- struct{}{}:
    default: // évite de bloquer si un flush est déjà en attente
    }
}

// flusher écoute les flushs et envoie un paquet unique
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
                // log ou ignore
                continue
            }
        }

        size := uint32(buf.Len())
        sizeBuf := make([]byte, 4)
        binary.BigEndian.PutUint32(sizeBuf, size)

        r.conn.Write(append(sizeBuf, buf.Bytes()...))
        r.buffer = nil // clear buffer
        r.mu.Unlock()
    }
}