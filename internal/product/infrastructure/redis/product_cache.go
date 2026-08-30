package redis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type ProductCache struct {
	address  string
	password string
	db       int
	timeout  time.Duration
	ttl      time.Duration
}

func NewProductCache(address, password string, db int, timeout, ttl time.Duration) *ProductCache {
	return &ProductCache{address: address, password: password, db: db, timeout: timeout, ttl: ttl}
}

func (c *ProductCache) Get(ctx context.Context, key string) ([]byte, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := c.prepare(conn); err != nil {
		return nil, err
	}
	if err := writeCommand(conn, "GET", key); err != nil {
		return nil, err
	}
	return readBulk(conn)
}

func (c *ProductCache) Set(ctx context.Context, key string, value []byte) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := c.prepare(conn); err != nil {
		return err
	}
	ttl := int64(c.ttl / time.Second)
	if ttl < 1 {
		ttl = 1
	}
	if err := writeCommand(conn, "SET", key, string(value), "EX", strconv.FormatInt(ttl, 10)); err != nil {
		return err
	}
	return readStatus(conn)
}

func (c *ProductCache) Delete(ctx context.Context, key string) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := c.prepare(conn); err != nil {
		return err
	}
	if err := writeCommand(conn, "DEL", key); err != nil {
		return err
	}
	_, err = readInteger(conn)
	return err
}

func (c *ProductCache) connect(ctx context.Context) (net.Conn, error) {
	timeout := c.timeout
	if timeout <= 0 {
		timeout = time.Second
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *ProductCache) prepare(conn net.Conn) error {
	if c.password != "" {
		if err := writeCommand(conn, "AUTH", c.password); err != nil {
			return err
		}
		if err := readStatus(conn); err != nil {
			return err
		}
	}
	if c.db != 0 {
		if err := writeCommand(conn, "SELECT", strconv.Itoa(c.db)); err != nil {
			return err
		}
		if err := readStatus(conn); err != nil {
			return err
		}
	}
	return nil
}

func writeCommand(w io.Writer, args ...string) error {
	if _, err := fmt.Fprintf(w, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len([]byte(arg)), arg); err != nil {
			return err
		}
	}
	return nil
}

func readLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errors.New("invalid redis response")
	}
	return line[:len(line)-2], nil
}

func readStatus(conn net.Conn) error {
	r := bufio.NewReader(conn)
	line, err := readLine(r)
	if err != nil {
		return err
	}
	if len(line) == 0 {
		return errors.New("empty redis response")
	}
	if line[0] == '-' {
		return errors.New(string(line[1:]))
	}
	if line[0] != '+' {
		return errors.New("unexpected redis status response")
	}
	return nil
}

func readInteger(conn net.Conn) (int64, error) {
	r := bufio.NewReader(conn)
	line, err := readLine(r)
	if err != nil {
		return 0, err
	}
	if len(line) == 0 || line[0] != ':' {
		return 0, errors.New("unexpected redis integer response")
	}
	return strconv.ParseInt(string(line[1:]), 10, 64)
}

func readBulk(conn net.Conn) ([]byte, error) {
	r := bufio.NewReader(conn)
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("invalid redis bulk response")
	}
	if line[0] == '-' {
		return nil, errors.New(string(line[1:]))
	}
	if line[0] != '$' {
		return nil, errors.New("unexpected redis bulk response")
	}
	n, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, errors.New("cache miss")
	}
	if n < -1 {
		return nil, errors.New("invalid redis bulk length")
	}
	value := make([]byte, n+2)
	if _, err := io.ReadFull(r, value); err != nil {
		return nil, err
	}
	if value[n] != '\r' || value[n+1] != '\n' {
		return nil, errors.New("invalid redis bulk payload")
	}
	return value[:n], nil
}
