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
	headerBuf := make([]byte, 16)
	if _, err := io.ReadFull(*r.conn, headerBuf); err != nil {
		return fmt.Errorf("failed to read packet header: %w", err)
	}

	magic := binary.BigEndian.Uint32(headerBuf[0:4])
	if magic != MAGIC_NUMBER {
		return fmt.Errorf("protocol corruption: invalid magic number (got 0x%x)", magic)
	}

	size := binary.BigEndian.Uint64(headerBuf[4:12])
	isError := binary.BigEndian.Uint32(headerBuf[12:16])

	data := make([]byte, size)
	if _, err := io.ReadFull(*r.conn, data); err != nil {
		return fmt.Errorf("failed to read packet payload: %w", err)
	}

	if isError != 0 {
		return fmt.Errorf("remote error: %s", string(data))
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	for _, d := range dests {
		if err := dec.Decode(d); err != nil {
			return fmt.Errorf("failed to decode gob data: %w", err)
		}
	}

	if buf.Len() != 0 {
		return fmt.Errorf("buffer has remaining unread data (%d bytes)", buf.Len())
	}
	return nil
}