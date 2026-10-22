package provicol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
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
	if t.NumOut() != 1 || !t.Out(0).Implements(reflect.TypeFor[error]()) {
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
		payload, file, err := readPacket(l.conn, 8)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(payload) < 1 {
			if file != nil {
				_ = file.Close()
			}
			return fmt.Errorf("request payload is empty")
		}

		opcode := askingBytecode(payload[0])
		payload = payload[1:]

		if cb, ok := l.megaMap[opcode]; ok {
			buf := bytes.NewBuffer(payload)
			dec := gob.NewDecoder(buf)

			var args []any
			if err := dec.Decode(&args); err != nil {
				if file != nil {
					_ = file.Close()
				}
				return err
			}

			if file != nil {
				args = append(args, file)
			}
			err := callUserFunction(cb, args)
			if file != nil {
				_ = file.Close()
			}
			if err != nil {
				l.responder.err = err
			}
			l.responder.flush()
			if err != nil {
				return err
			}
		} else if file != nil {
			_ = file.Close()
		}
	}
}

func (c *Child) Reply(v any) {
	c.responder.reply(v)
}

// ReplyFD attaches a file descriptor to the next response.
func (c *Child) ReplyFD(file *os.File) error {
	return c.responder.replyFD(file)
}

func (c *Child) Error(what error) {
	c.responder.throw(what)
}

func (c *Child) Flush() {
	c.responder.flush()
}

func (c *Child) Close() error {
	return c.conn.Close()
}
