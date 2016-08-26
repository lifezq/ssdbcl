// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package ssdbcl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Config struct {
	Host    string
	Port    uint16
	Auth    string
	Timeout uint8
}

type Client struct {
	sock *net.TCPConn
}

type KeyValue struct {
	Key   string
	Value string
}

type Reply struct {
	State string
	Data  []string
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

func (c *Client) Cmd(args ...interface{}) *Reply {

	reply := &Reply{
		State: "not_found",
		Data:  []string{},
	}

	if err := c.send(args); err != nil {
		return reply
	}

	resp, err := c.recv()
	if err != nil {
		return reply
	}

	for k, s := range resp {

		if k == 0 {
			reply.State = s
			continue
		}

		reply.Data = append(reply.Data, s)
	}

	return reply
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

		case [][]byte:
			for _, bs := range arg {
				buf.WriteString(fmt.Sprintf("%d", len(s)))
				buf.WriteByte('\n')
				buf.WriteString(string(bs))
				buf.WriteByte('\n')
			}
			continue

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
		case uint:
			s = fmt.Sprintf("%d", arg)

		case int8:
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
		buf  [8192]byte
		idx  = 0
		resp []string
	)

	for {

		n, err := c.sock.Read(buf[0:])
		if err != nil {
			return nil, err
		}

		offset := 0

		for {

			idx = bytes.IndexByte(buf[offset:], '\n')
			if idx == -1 {
				//				log.Printf("----------------------------------idx:%d\n", idx)
				break
			}

			p := buf[offset : offset+idx]

			//log.Printf("*****************************idx:%d--p:%s#########\n", idx, string(p))

			offset += idx + 1

			if len(p) == 0 || (len(p) == 1 && p[0] == '\r') {

				if len(resp) > 0 {
					return resp, nil
				}

				continue
			}

			if size, err := strconv.Atoi(string(p)); err != nil || size < 1 {
				resp = append(resp, string(p))
				continue
			}

			if offset >= n {
				break
			}
		}
	}

	return resp, nil
}

func (r *Reply) StateString() string {
	return r.State
}

func (r *Reply) Bytes() []byte {

	if len(r.Data) < 1 {
		return []byte{}
	}

	return []byte(r.Data[0])
}

func (r *Reply) String() string {

	if len(r.Data) < 1 {
		return ""
	}

	return r.Data[0]
}

func (r *Reply) List() []string {
	return r.Data
}

func (r *Reply) Hash() []KeyValue {

	if len(r.Data) < 2 {
		return []KeyValue{}
	}

	kvs := []KeyValue{}

	dlen := len(r.Data)
	for i := 0; i < dlen; i += 2 {

		kvs = append(kvs, KeyValue{
			Key:   r.Data[i],
			Value: r.Data[i+1],
		})
	}

	return kvs
}

func (r *Reply) ReplyJson(v interface{}) error {

	defer func() {

		if r := recover(); r != nil {
			return
		}
	}()

	if len(r.Data) < 1 {
		return fmt.Errorf("Not Found")
	}

	return json.Unmarshal([]byte(r.Data[0]), &v)
}
