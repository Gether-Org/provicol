package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

type Sender struct {
    conn net.Conn
}

func SenderWithPath(path string) (*Sender, error) {
    c, err := net.Dial("unix", path)
    if err != nil {
        return nil, err
    }
    return &Sender{conn: c}, nil
}

func (m *Sender) Ask(op askingBytecode, args ...any) *ResponseScanner {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    if len(args) > 0 {
        enc.Encode(args[0]) // simplifié : un argument
    }

    size := uint32(buf.Len() + 1)
    sizeBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(sizeBuf, size)

    m.conn.Write(append(sizeBuf, append([]byte{byte(op)}, buf.Bytes()...)...))
    return &ResponseScanner{conn: m.conn}
}

type ResponseScanner struct {
    conn net.Conn
}

func (r *ResponseScanner) Scan(dest any) error {
    sizeBuf := make([]byte, 4)
    if _, err := io.ReadFull(r.conn, sizeBuf); err != nil {
        return err
    }
    size := binary.BigEndian.Uint32(sizeBuf)

    data := make([]byte, size)
    if _, err := io.ReadFull(r.conn, data); err != nil {
        return err
    }

    buf := bytes.NewBuffer(data)
    dec := gob.NewDecoder(buf)
    return dec.Decode(dest)
}
