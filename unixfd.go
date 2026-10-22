package provicol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
)

const maxFrameSize = 1 << 30

func writePacket(conn net.Conn, packet []byte, file *os.File) error {
	if file == nil {
		n, err := conn.Write(packet)
		if err != nil {
			return err
		}
		if n != len(packet) {
			return io.ErrShortWrite
		}
		return nil
	}

	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("file descriptor transfer requires a Unix connection")
	}

	rights := syscall.UnixRights(int(file.Fd()))
	n, _, err := unixConn.WriteMsgUnix(packet, rights, nil)
	if err != nil {
		return err
	}
	if n != len(packet) {
		return io.ErrShortWrite
	}
	return nil
}

func readPacket(conn net.Conn, headerSize int) ([]byte, *os.File, error) {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return readPacketWithoutFD(conn, headerSize)
	}

	header := make([]byte, headerSize)
	oob := make([]byte, syscall.CmsgSpace(4))
	n, oobn, _, _, err := unixConn.ReadMsgUnix(header, oob)
	if err != nil {
		return nil, nil, err
	}
	if n < headerSize {
		if _, err := io.ReadFull(conn, header[n:]); err != nil {
			return nil, nil, err
		}
	}

	var file *os.File
	if oobn > 0 {
		messages, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil {
			return nil, nil, fmt.Errorf("parse file descriptor message: %w", err)
		}
		for _, message := range messages {
			fds, err := syscall.ParseUnixRights(&message)
			if err != nil {
				return nil, nil, fmt.Errorf("parse received file descriptor: %w", err)
			}
			for _, fd := range fds {
				if file != nil {
					_ = syscall.Close(fd)
					return nil, nil, fmt.Errorf("received more than one file descriptor")
				}
				file = os.NewFile(uintptr(fd), "provicol-transfer")
			}
		}
	}

	sizeOffset := 0
	if headerSize == 16 {
		sizeOffset = 4
	}
	size := binary.BigEndian.Uint64(header[sizeOffset : sizeOffset+8])
	if size > maxFrameSize {
		if file != nil {
			_ = file.Close()
		}
		return nil, nil, fmt.Errorf("frame size %d exceeds limit", size)
	}
	if headerSize == 16 {
		magic := binary.BigEndian.Uint32(header[0:4])
		if magic != magicNumber {
			if file != nil {
				_ = file.Close()
			}
			return nil, nil, fmt.Errorf("protocol corruption: invalid magic number (got 0x%x)", magic)
		}
		if binary.BigEndian.Uint32(header[12:16]) != 0 {
			if file != nil {
				_ = file.Close()
			}
			payload := make([]byte, size)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return nil, nil, err
			}
			return nil, nil, fmt.Errorf("remote error: %s", string(payload))
		}
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(conn, payload); err != nil {
		if file != nil {
			_ = file.Close()
		}
		return nil, nil, err
	}
	return payload, file, nil
}

func readPacketWithoutFD(conn net.Conn, headerSize int) ([]byte, *os.File, error) {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, nil, err
	}
	sizeOffset := 0
	if headerSize == 16 {
		sizeOffset = 4
	}
	size := binary.BigEndian.Uint64(header[sizeOffset : sizeOffset+8])
	if size > maxFrameSize {
		return nil, nil, fmt.Errorf("frame size %d exceeds limit", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, nil, err
	}
	return payload, nil, nil
}
