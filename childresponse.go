package provicol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
    "fmt"
)

type ChildResponse struct {
    conn *net.Conn
}

func (r *ChildResponse) Scan(dests ...any) error {
    sizeBuf := make([]byte, 4)
    if _, err := io.ReadFull(*r.conn, sizeBuf); err != nil {
        return err
    }
    size := binary.BigEndian.Uint32(sizeBuf)

    data := make([]byte, size)
    if _, err := io.ReadFull(*r.conn, data); err != nil {
        return err
    }

    buf := bytes.NewBuffer(data)
    dec := gob.NewDecoder(buf)

    for _, d := range dests {
        if err := dec.Decode(d); err != nil {
            return err
        }
    }

    if buf.Len() != 0 {
        return fmt.Errorf("buffer has remaining data (%d bytes)", buf.Len())
    }
    return nil
}