// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package ssdbcl

import (
	"fmt"
	"testing"
)

func conn() (*Client, error) {

	c, err := New(Config{
		Host:    "127.0.0.1",
		Port:    6380,
		Auth:    "",
		Timeout: 3,
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

		if c.Cmd("del", k).StateString() != "ok" {
			t.Errorf("del error:%d\n", i)
		}
	}
}
