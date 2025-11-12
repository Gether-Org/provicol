package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
	"os"
)

type Parent struct {
    io.Closer

    sock net.Listener
	conn net.Conn
}

func NewParent(socketPath string) (*Parent, error) {
    var err error
    p := &Parent{}
    _ = os.Remove(socketPath)

    p.sock, err = net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	os.Chmod(socketPath, 0777)

	p.conn, err = p.sock.Accept()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (m *Parent) Ask(op askingBytecode, args ...any) *ChildResponse {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if len(args) > 0 {
		enc.Encode(args[0])
	}

	size := uint32(buf.Len() + 1)
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, size)

	m.conn.Write(append(sizeBuf, append([]byte{byte(op)}, buf.Bytes()...)...))
	return &ChildResponse{conn: m.conn}
}

func (m *Parent) Close() error {
    err := m.sock.Close()
    if err != nil {
        return err
    }
    return nil
}
