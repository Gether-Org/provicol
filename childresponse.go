package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

type ChildResponse struct {
    conn net.Conn
}

func (r *ChildResponse) Scan(dest any) error {
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
