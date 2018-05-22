// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lifezq/ssdbcl"
)

func main() {

	c, err := ssdbcl.New(&ssdbcl.Config{
		Host: "127.0.0.1",
		Port: 6380,
		Auth: "auth_string",
	})

	defer c.Close()
	if err != nil {
		log.Printf("conn err:%s\n", err.Error())
		return
	}

	ks, kvs := []string{}, []string{}
	for i := 1; i < 100; i++ {
		ks = append(ks, fmt.Sprintf("example_test.%d", i))
		kvs = append(kvs, fmt.Sprintf("example_test.%d", i))
		kvs = append(kvs, fmt.Sprintf("example_value.%d", i))
	}

	if rs := c.Cmd("multi_set", kvs); rs.State != ssdbcl.ReplyOK {
		log.Printf("multi_set error state::%s\n", rs.State)
	}

	rs := c.Cmd("multi_get", ks).Hash()
	for _, v := range rs {
		log.Printf("v.key:%s v.value:%s\n", v.Key, v.Value)
	}

	if rs := c.Cmd("multi_del", ks); rs.State != ssdbcl.ReplyOK {
		log.Printf("multi_del error:%d\n", rs.State)
	}

	//
	c.Cmd("set", "test.0", "val.0")

	//
	rsp := c.Cmd("get", "test.0")
	log.Printf("**get** state:%s rsp:%v\n", rsp.State, rsp.String())

	//
	c.Cmd("del", "test.0")

	rsp = c.Cmd("scan", "0", "z", 10)
	log.Printf("**scan** state:%s rsp:%v\n", rsp.State, rsp.Hash())
	// etc...

	time.Sleep(2e9)
}
