package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
	"os"
	"sync"
)

type Parent struct {
	io.Closer

	sock    net.Listener
	conn    net.Conn
	writeMu sync.Mutex
}

func NewParent(socketPath string, perms os.FileMode) (p *Parent, err error) {
	p = &Parent{}
	err = nil
	_ = os.Remove(socketPath)

	p.sock, err = net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(socketPath, perms); err != nil {
		return nil, err
	}

	p.conn, err = p.sock.Accept()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (m *Parent) Ask(op askingBytecode, args ...any) *ChildResponse {
	return m.ask(nil, op, args...)
}

// AskWithFD sends a request together with a file descriptor. The child
// receives the descriptor as the last argument of its bound handler.
func (m *Parent) AskWithFD(file *os.File, op askingBytecode, args ...any) *ChildResponse {
	return m.ask(file, op, args...)
}

func (m *Parent) ask(file *os.File, op askingBytecode, args ...any) *ChildResponse {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(args); err != nil {
		panic(err)
	}

	size := uint64(buf.Len() + 1)
	sizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuf, size)

	packet := append(sizeBuf, append([]byte{byte(op)}, buf.Bytes()...)...)
	m.writeMu.Lock()
	err := writePacket(m.conn, packet, file)
	m.writeMu.Unlock()
	if err != nil {
		panic(err)
	}
	return &ChildResponse{conn: &m.conn}
}

func (m *Parent) Close() error {
	err := m.sock.Close()
	if err != nil {
		return err
	}
	return nil
}
