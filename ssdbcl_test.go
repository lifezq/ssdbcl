// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package ssdbcl

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func conn() (*Client, error) {

	c, err := New(&Config{
		Host: "127.0.0.1",
		Port: 8888,
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

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {

				k, v := fmt.Sprintf("test.%d", i), fmt.Sprintf("val.%d", i)

				rp := c.Cmd("set", k, v)

				if gv := c.Cmd("get", k).String(); gv != v {
					t.Errorf("get error:%d set:%s got:%s rp:%v\n", i, v, gv, rp)
				}
			}
		}()
	}

	wg.Wait()
}

func TestCmdDel(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {

		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 1000; i++ {

				k := fmt.Sprintf("test.%d", i)

				if c.Cmd("del", k).State != "ok" {
					t.Errorf("del error:%d\n", i)
				}
			}
		}()
	}

	wg.Wait()
}

func TestCmdScan(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {

		wg.Add(1)
		go func() {
			defer wg.Done()

			ks := []string{}

			for i := 0; i < 1000; i++ {

				k, v := fmt.Sprintf("test.%d", i), fmt.Sprintf("vall.%d", i)

				if c.Cmd("set", k, v).State != ReplyOK {
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

			if rs := c.Cmd("multi_del", ks); rs.State != ReplyOK {
				t.Errorf("scan.multi_del error:%s\n", rs.State)
			}

		}()
	}

	wg.Wait()
}

func TestCmdMultiSetGetDel(t *testing.T) {

	c, err := conn()
	defer c.Close()

	if err != nil {
		t.Fatalf("conn:%s\n", err.Error())
	}

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {

		wg.Add(1)
		go func() {
			defer wg.Done()

			ks := []string{}

			ks, kvs := []string{}, []string{}

			for i := 0; i < 1000; i++ {

				ks = append(ks, fmt.Sprintf("multi_tests.%d", i))

				kvs = append(kvs, fmt.Sprintf("multi_tests.%d", i))
				kvs = append(kvs, fmt.Sprintf("multi_value.%d", i))
			}

			if rs := c.Cmd("multi_set", kvs); rs.State != ReplyOK {
				t.Errorf("multi_set error:%s\n", rs.State)
			}

			rs := c.Cmd("multi_get", ks).Hash()
			for _, v := range rs {

				idx := strings.LastIndex(v.Key, ".")
				if v.Key[idx:] != v.Value[idx:] {
					t.Errorf("multi_get error v.Key:%s  v.Value:%s\n", v.Key[idx:], v.Value[idx:])
				}
			}

			if rs := c.Cmd("multi_del", ks); rs.State != ReplyOK {
				t.Errorf("multi_del error:%s\n", rs.State)
			}

		}()
	}

	wg.Wait()
}
