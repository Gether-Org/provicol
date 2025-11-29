package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"reflect"
)

type Child struct {
	conn      net.Conn
	megaMap   map[askingBytecode]any
	responder *responder
}

func NewChild(socketPath string) (*Child, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	child := &Child{
		conn:      conn,
		megaMap:   make(map[askingBytecode]any),
		responder: newResponder(&conn),
	}
	return child, nil
}

func (l *Child) Bind(op askingBytecode, fn any) {
	if l.megaMap == nil {
		l.megaMap = make(map[askingBytecode]any)
	}
	l.megaMap[op] = fn
}

func callUserFunction(fn any, args []any) error {
    v := reflect.ValueOf(fn)
    t := v.Type()

    if t.Kind() != reflect.Func {
        return fmt.Errorf("Bind: handler must be a function")
    }
    if t.NumOut() != 1 || !t.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
        return fmt.Errorf("Bind: handler must return exactly 1 error")
    }
    if len(args) != t.NumIn() {
        return fmt.Errorf("Bind: expected %d args, got %d", t.NumIn(), len(args))
    }

    callArgs := make([]reflect.Value, len(args))
    for i := 0; i < t.NumIn(); i++ {
        expected := t.In(i)
        got := reflect.ValueOf(args[i])
        if !got.Type().AssignableTo(expected) {
            return fmt.Errorf(
                "Bind: argument %d wrong type: expected %s, got %s",
                i, expected, got.Type(),
            )
        }
        callArgs[i] = got
    }

    out := v.Call(callArgs)
    if !out[0].IsNil() {
        return out[0].Interface().(error)
    }
    return nil
}


func (l *Child) Listen() error {
	for {
		sizeBuf := make([]byte, 8)
		if _, err := io.ReadFull(l.conn, sizeBuf); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		size := binary.BigEndian.Uint64(sizeBuf)

		data := make([]byte, size)
		if _, err := io.ReadFull(l.conn, data); err != nil {
			return err
		}

		opcode := askingBytecode(data[0])
		payload := data[1:]

		if cb, ok := l.megaMap[opcode]; ok {
			buf := bytes.NewBuffer(payload)
			dec := gob.NewDecoder(buf)

			var args []any
			if err := dec.Decode(&args); err != nil {
				return err
			}

			if err := callUserFunction(cb, args); err != nil {
				return err
			}
			l.responder.flush()
		}
	}
}

func (c *Child) Reply(v any) {
    c.responder.reply(v)
}

func (c *Child) Flush() {
    c.responder.flush()
}

func (c *Child) Close() error {
	return c.conn.Close()
}
