package provicol

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileDescriptorRoundTrip(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "provicol.sock")
	sourcePath := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(sourcePath, []byte("hello through fd"), 0600); err != nil {
		t.Fatal(err)
	}

	parentCh := make(chan *Parent, 1)
	errCh := make(chan error, 1)
	go func() {
		parent, err := NewParent(socketPath, 0600)
		if err != nil {
			errCh <- err
			return
		}
		parentCh <- parent
	}()

	var child *Child
	var err error
	for range 20 {
		child, err = NewChild(socketPath)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer child.Close()

	parent := <-parentCh
	defer parent.Close()

	child.Bind(PutObject, func(file *os.File) error {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if string(data) != "hello through fd" {
			t.Errorf("received %q", data)
		}
		child.Reply("ok")
		return nil
	})
	child.Bind(GetObject, func() error {
		file, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer file.Close()
		return child.ReplyFD(file)
	})

	listenErr := make(chan error, 1)
	go func() { listenErr <- child.Listen() }()

	source, err := os.Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	response := parent.AskWithFD(source, PutObject)
	_ = source.Close()
	var status string
	if err := response.Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "ok" {
		t.Fatalf("status = %q", status)
	}

	file, err := parent.Ask(GetObject).ScanFD()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello through fd" {
		t.Fatalf("downloaded %q", data)
	}
}
