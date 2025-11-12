package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

type middleWareCallBack func(*Responder, ...any) error

type Child struct {
	conn      net.Conn
	megaMap   map[askingBytecode]middleWareCallBack
	responder *Responder
}

func NewChild(socketPath string) (*Child, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	child := &Child{
		conn:    conn,
		megaMap: make(map[askingBytecode]middleWareCallBack),
		responder: newResponder(&conn),
	}
	return child, nil
}

func (l *Child) Bind(op askingBytecode, f middleWareCallBack) {
	if l.megaMap == nil {
		l.megaMap = make(map[askingBytecode]middleWareCallBack)
	}
	l.megaMap[op] = f
}

func (l *Child) Listen() error {
	for {
		sizeBuf := make([]byte, 4)
		if _, err := io.ReadFull(l.conn, sizeBuf); err != nil {
			if err == io.EOF {
				return nil
			}
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
				var arg any
				if err := dec.Decode(&arg); err == nil {
					args = append(args, arg)
				}
			}

			if err := cb(l.responder, args...); err != nil {
				return err
			}

			l.responder.Flush()
		}
	}
}

