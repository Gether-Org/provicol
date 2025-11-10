package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

type middleWareCallBack func(*Responder, ...any) error

type Listener struct {
    conn     net.Conn
    megaMap  map[askingBytecode]middleWareCallBack
	responder *Responder
}

func ListenerWithPath(socketPath string) (*Listener, error) {
    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return nil, err
    }

    listener := &Listener{
        conn:    conn,
        megaMap: make(map[askingBytecode]middleWareCallBack),
        responder: &Responder{
            conn:   conn,
            buffer: make([]any, 0),
            sendCh: make(chan struct{}, 1),
        },
    }

    return listener, nil
}

func (l *Listener) Bind(op askingBytecode, f middleWareCallBack) {
    if l.megaMap == nil {
        l.megaMap = make(map[askingBytecode]middleWareCallBack)
    }
    l.megaMap[op] = f
}

func (l *Listener) Listen() error {
    for {
        sizeBuf := make([]byte, 4)
        if _, err := io.ReadFull(l.conn, sizeBuf); err != nil {
            return err
        }
        size := binary.BigEndian.Uint32(sizeBuf)

        data := make([]byte, size)
        if _, err := io.ReadFull(l.conn, data); err != nil {
            return err
        }

        opcode := askingBytecode(data[0])
        payload := data[1:]

        if cb, ok := l.megaMap[opcode]; ok {
            var args []any
            if len(payload) > 0 {
                buf := bytes.NewBuffer(payload)
                dec := gob.NewDecoder(buf)
                // pour simplifier, on peut toujours décoder un seul argument ici
                var arg any
                if err := dec.Decode(&arg); err == nil {
                    args = append(args, arg)
                }
            }
            cb(l.responder, args...)
        }
    }
}
