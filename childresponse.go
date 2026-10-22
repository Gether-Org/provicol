package provicol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

type ChildResponse struct {
	conn *net.Conn
}

func (r *ChildResponse) Scan(dests ...any) error {
	data, file, err := r.read()
	if err != nil {
		return err
	}
	if file != nil {
		_ = file.Close()
		return fmt.Errorf("response unexpectedly contains a file descriptor")
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

// ScanFD returns the file descriptor attached to a successful response.
func (r *ChildResponse) ScanFD() (*os.File, error) {
	data, file, err := r.read()
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("response does not contain a file descriptor")
	}
	if len(data) != 0 {
		_ = file.Close()
		return nil, fmt.Errorf("file descriptor response contains an unexpected payload")
	}
	return file, nil
}

func (r *ChildResponse) read() ([]byte, *os.File, error) {
	data, file, err := readPacket(*r.conn, 16)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}
	return data, file, nil
}
