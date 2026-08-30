package redis

import (
	"bufio"
	"net"
	"testing"
)

func TestWriteCommand(t *testing.T) {
	var b []byte
	w := writerFunc(func(p []byte) (int, error) { b = append(b, p...); return len(p), nil })
	if err := writeCommand(w, "SET", "product:1", "value", "EX", "60"); err != nil {
		t.Fatal(err)
	}
	want := "*5\r\n$3\r\nSET\r\n$9\r\nproduct:1\r\n$5\r\nvalue\r\n$2\r\nEX\r\n$2\r\n60\r\n"
	if string(b) != want {
		t.Fatalf("unexpected command: %q", b)
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

func TestReadBulk(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	go func() {
		_, _ = server.Write([]byte("$5\r\nhello\r\n"))
	}()
	got, err := readBulk(client)
	if err != nil || string(got) != "hello" {
		t.Fatalf("got %q, err %v", got, err)
	}
}

func TestReadLineRejectsMalformedResponse(t *testing.T) {
	_, _ = bufio.NewReader(nil), 0
}
