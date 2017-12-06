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

const (
	ReplyOK = "ok"
)

type Config struct {
	Host string
	Port uint16
	Auth string
}

type Client struct {
	sock     *net.TCPConn
	recv_buf bytes.Buffer
	cfg      *Config
	cmdchan  chan bool
}

type KeyValue struct {
	Key   string
	Value string
}

type Reply struct {
	State string
	Data  []string
}

func New(c *Config) (*Client, error) {

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	cl := &Client{
		sock:    conn,
		cfg:     c,
		cmdchan: make(chan bool, 1),
	}

	if len(c.Auth) > 0 {

		if auth_rs_state := cl.Cmd("auth", c.Auth).State; auth_rs_state != ReplyOK {
			cl.Close()
			return cl, fmt.Errorf("auth error:%s", auth_rs_state)
		}
	}

	return cl, nil
}

func (c *Client) Close() {

	if c != nil && c.sock != nil {
		c.sock.Close()
		c.sock = nil
	}
}

func (c *Client) Cmd(args ...interface{}) *Reply {

	c.cmdchan <- false

	reply := &Reply{
		State: "client_error",
		Data:  []string{},
	}

	if c.sock == nil {
		<-c.cmdchan
		return reply
	}

	for redo := 1; redo <= 3; redo++ {

		if redo == 3 {
			c.Close()
			if conn, err := New(c.cfg); err == nil {
				*c = *conn
			}
		}

		c.sock.SetDeadline(time.Now().Add(3 * time.Second))

		if err := c.send(args); err != nil {
			time.Sleep(time.Duration(redo*100) * time.Millisecond)
			continue
		}

		resp, err := c.recv()
		if err != nil {
			time.Sleep(time.Duration(redo*100) * time.Millisecond)
			continue
		}

		for k, s := range resp {
			if k == 0 {
				reply.State = s
				continue
			}

			reply.Data = append(reply.Data, s)
		}

		break
	}

	<-c.cmdchan
	return reply
}

func (c *Client) send(args []interface{}) error {

	var buf bytes.Buffer

	for _, arg := range args {

		s := ""

		switch arg := arg.(type) {
		case byte:
			s = fmt.Sprintf("%d", arg)

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

	var tmp [8192]byte

	for {

		resp := c.parse()
		if resp == nil || len(resp) > 0 {
			return resp, nil
		}

		n, err := c.sock.Read(tmp[0:])
		if err != nil {
			return nil, err
		}

		c.recv_buf.Write(tmp[0:n])
	}
}

func (c *Client) parse() []string {

	var (
		idx    = 0
		offset = 0
		resp   []string
		buf    = c.recv_buf.Bytes()
	)

	for {

		idx = bytes.IndexByte(buf[offset:], '\n')
		if idx == -1 {
			// log.Printf("---------idx:%d---buf:%s\n", idx, string(buf[offset:]))
			break
		}

		p := buf[offset : offset+idx]

		offset += idx + 1
		//log.Printf("********idx:%d--p:%s#########\n", idx, string(p))

		if len(p) == 0 || (len(p) == 1 && p[0] == '\r') {

			if len(resp) == 0 {
				continue
			} else {
				c.recv_buf.Next(offset)
				return resp
			}
		}

		size, err := strconv.Atoi(string(p))
		if err != nil || size < 0 {
			return nil
		}

		if offset+size >= c.recv_buf.Len() {
			break
		}

		v := buf[offset : offset+size]

		resp = append(resp, string(v))

		offset += size + 1
	}

	return []string{}
}

func (r *Reply) Int() int {

	if len(r.Data) < 1 {
		return 0
	}

	i, _ := strconv.Atoi(r.Data[0])
	return i
}

func (r *Reply) Int32() int32 {

	if len(r.Data) < 1 {
		return 0
	}

	i, _ := strconv.Atoi(r.Data[0])
	return int32(i)
}

func (r *Reply) Int64() int64 {

	if len(r.Data) < 1 {
		return 0
	}

	i, _ := strconv.ParseInt(r.Data[0], 10, 64)
	return i
}

func (r *Reply) Uint64() uint64 {

	if len(r.Data) < 1 {
		return 0
	}

	i, _ := strconv.ParseUint(r.Data[0], 10, 64)
	return i
}

func (r *Reply) Float64() float64 {

	if len(r.Data) < 1 {
		return 0
	}

	i, _ := strconv.ParseFloat(r.Data[0], 64)
	return i
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

func (r *Reply) Bool() bool {

	if len(r.Data) < 1 {
		return false
	}

	if b, err := strconv.ParseBool(r.Data[0]); err == nil {
		return b
	}

	return false
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
	for i := 0; i < dlen-1; i += 2 {

		kvs = append(kvs, KeyValue{
			Key:   r.Data[i],
			Value: r.Data[i+1],
		})
	}

	return kvs
}

func (r *Reply) ReplyJson(v interface{}) error {

	defer recover()

	if len(r.Data) < 1 {
		return fmt.Errorf("Not Found")
	}

	return json.Unmarshal([]byte(r.Data[0]), &v)
}

func (r *Reply) ReplyOk() bool {
	return r.State == ReplyOK
}

func (kv KeyValue) ReplyJson(v interface{}) error {

	defer recover()

	if kv.Key == "" {
		return fmt.Errorf("Not Found")
	}

	return json.Unmarshal([]byte(kv.Value), &v)
}
