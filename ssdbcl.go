package ssdbcl

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	sock *net.TCPConn
}

func New(c Config) (*Client, error) {

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	if c.Timeout < 1 {
		c.Timeout = 3
	}

	conn.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Timeout)))

	return &Client{
		sock: conn,
	}, nil
}

func (c *Client) Close() error {
	return c.sock.Close()
}

func (c *Client) Do(args ...interface{}) ([]string, error) {
	err := c.send(args)
	if err != nil {
		return []string{}, err
	}

	return c.recv()
}

func (c *Client) send(args []interface{}) error {

	var buf bytes.Buffer

	for _, arg := range args {

		s := ""

		switch arg := arg.(type) {
		case byte:
			s = fmt.Sprintf("%c", arg)

		case []byte:
			s = string(arg)

		case string:
			s = arg

		case []string:
			for _, s := range arg {
				buf.WriteString(fmt.Sprintf("%d", len(s)))
				buf.WriteByte('\n')
				buf.WriteString(s)
				buf.WriteByte('\n')
			}
			continue

		case int:
			s = fmt.Sprintf("%d", arg)

		case int32:
			s = fmt.Sprintf("%d", arg)

		case uint32:
			s = fmt.Sprintf("%d", arg)

		case int64:
			s = fmt.Sprintf("%d", arg)

		case uint64:
			s = fmt.Sprintf("%d", arg)

		case float64:
			s = fmt.Sprintf("%f", arg)

		case bool:
			if arg {
				s = "1"
			} else {
				s = "0"
			}

		case nil:
			s = ""

		default:
			return fmt.Errorf("bad arguments")
		}

		buf.WriteString(fmt.Sprintf("%d", len(s)))
		buf.WriteByte('\n')
		buf.WriteString(s)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')

	_, err := c.sock.Write(buf.Bytes())
	return err
}

func (c *Client) recv() ([]string, error) {

	var (
		r      [8192]byte
		offset = 0
	)

	for {

		n, err := c.sock.Read(r)
		if err != nil {
			return nil, err
		}

		for {
			bytes.IndexByte(r[offset:], '\n')
		}
	}
}

func (c *Client) Get(key string) ([]string, error) {

	n, err := c.Do("get", key)

	return n, err
}
