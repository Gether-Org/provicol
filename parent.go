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

func NewParent(socketPath string, perms os.FileMode) (*Parent, error) {
    var err error
    p := &Parent{}
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
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)

    if err := enc.Encode(args); err != nil {
        panic(err)
    }

    size := uint64(buf.Len() + 1)
    sizeBuf := make([]byte, 8)
    binary.BigEndian.PutUint64(sizeBuf, size)

    m.conn.Write(append(sizeBuf, append([]byte{byte(op)}, buf.Bytes()...)...))
    return &ChildResponse{conn: &m.conn}
}


func (m *Parent) Close() error {
    err := m.sock.Close()
    if err != nil {
        return err
    }
    return nil
}
