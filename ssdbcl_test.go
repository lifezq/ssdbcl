// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package ssdbcl

import (
	"fmt"
	"strings"
	"testing"
)

func conn() (*Client, error) {

	c, err := New(Config{
		Host: "127.0.0.1",
		Port: 6380,
		Auth: "",
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}

func TestCmdSetGet(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	for i := 0; i < 1000; i++ {

		k, v := fmt.Sprintf("test.%d", i), fmt.Sprintf("val.%d", i)

		c.Cmd("set", k, v)

		if c.Cmd("get", k).String() != v {
			t.Errorf("get error:%d\n", i)
		}
	}
}

func TestCmdDel(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	for i := 0; i < 1000; i++ {

		k := fmt.Sprintf("test.%d", i)

		if c.Cmd("del", k).State != "ok" {
			t.Errorf("del error:%d\n", i)
		}
	}
}

func TestCmdScan(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	ks := []string{}

	for i := 0; i < 1000; i++ {

		k, v := fmt.Sprintf("test.%d", i), fmt.Sprintf("vall.%d", i)

		if c.Cmd("set", k, v).State != ReplyOk {
			t.Errorf("scan.set error:%d\n", i)
		}

		ks = append(ks, k)
	}

	kvs := c.Cmd("scan", "test.0", "test.z", 100).Hash()
	for _, v := range kvs {

		if idx := strings.LastIndex(v.Key, "."); v.Key[idx:] != v.Value[idx:] {
			t.Errorf("scan error idx:%d k:%s v:%s\n", idx, v.Key[idx:], v.Value[idx:])
		}
	}

	if rs := c.Cmd("multi_del", ks); rs.State != ReplyOk {
		t.Errorf("scan.multi_del error:%d\n", rs.State)
	}
}

func TestCmdMultiSetGetDel(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	ks, kvs := []string{}, []string{}

	for i := 0; i < 1000; i++ {

		ks = append(ks, fmt.Sprintf("multi_tests.%d", i))

		kvs = append(kvs, fmt.Sprintf("multi_tests.%d", i))
		kvs = append(kvs, fmt.Sprintf("multi_value.%d", i))
	}

	if rs := c.Cmd("multi_set", kvs); rs.State != ReplyOk {
		t.Errorf("multi_set error:%d\n", rs.State)
	}

	rs := c.Cmd("multi_get", ks).Hash()
	for _, v := range rs {

		idx := strings.LastIndex(v.Key, ".")
		if v.Key[idx:] != v.Value[idx:] {
			t.Errorf("multi_get error v.Key:%s  v.Value:%s\n", v.Key[idx:], v.Value[idx:])
		}
	}

	if rs := c.Cmd("multi_del", ks); rs.State != ReplyOk {
		t.Errorf("multi_del error:%d\n", rs.State)
	}
}
