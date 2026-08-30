package redis

import (
	"bytes"
	"net"
	"testing"
)

func TestWriteCommand(t *testing.T) {
	var b bytes.Buffer
	if err := writeCommand(&b, "SET", "product:1", "value", "EX", "60"); err != nil { t.Fatal(err) }
	want := "*5\r\n$3\r\nSET\r\n$9\r\nproduct:1\r\n$5\r\nvalue\r\n$2\r\nEX\r\n$2\r\n60\r\n"
	if b.String() != want { t.Fatalf("unexpected command: %q", b.String()) }
}

func TestReadBulk(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close(); defer client.Close()
	go func() { _, _ = server.Write([]byte("$5\r\nhello\r\n")) }()
	got, err := readBulk(client)
	if err != nil || string(got) != "hello" { t.Fatalf("got %q, err %v", got, err) }
}

func TestReadBulkCacheMiss(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close(); defer client.Close()
	go func() { _, _ = server.Write([]byte("$-1\r\n")) }()
	if _, err := readBulk(client); err == nil { t.Fatal("expected cache miss") }
}
