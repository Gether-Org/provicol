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
	responder *Responder
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

func callUserFunction(fn any, responder *Responder, args []any) error {
	v := reflect.ValueOf(fn)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return fmt.Errorf("Bind: handler must be a function")
	}

	if t.NumIn() == 0 {
		return fmt.Errorf("Bind: handler must accept *Responder as first argument")
	}
	if t.In(0) != reflect.TypeOf((*Responder)(nil)) {
		return fmt.Errorf("Bind: first argument must be *Responder")
	}

	if len(args) != t.NumIn()-1 {
		return fmt.Errorf("Bind: expected %d args, got %d", t.NumIn()-1, len(args))
	}

	callArgs := make([]reflect.Value, 0, len(args)+1)
	callArgs = append(callArgs, reflect.ValueOf(responder))

	for i := 1; i < t.NumIn(); i++ {
		expected := t.In(i)
		got := reflect.ValueOf(args[i-1])

		if !got.Type().AssignableTo(expected) {
			return fmt.Errorf(
				"Bind: argument %d wrong type: expected %s, got %s",
				i, expected, got.Type(),
			)
		}

		callArgs = append(callArgs, got)
	}

	out := v.Call(callArgs)
	if len(out) != 1 {
		return fmt.Errorf("Bind: handler must return exactly 1 value (error)")
	}

	if !out[0].IsNil() {
		return out[0].Interface().(error)
	}
	return nil
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
			buf := bytes.NewBuffer(payload)
			dec := gob.NewDecoder(buf)

			var args []any
			if err := dec.Decode(&args); err != nil {
				return err
			}

			if err := callUserFunction(cb, l.responder, args); err != nil {
				return err
			}
			l.responder.Flush()
		}
	}
}

func (c *Child) Close() error {
	return c.conn.Close()
}
