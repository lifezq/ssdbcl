// Copyright 2016 The Ssdbcl Author. All Rights Reserved.

package main

import (
	"log"
	"time"

	"github.com/lifezq/ssdbcl"
)

func main() {

	c, err := ssdbcl.New(ssdbcl.Config{
		Host:    "127.0.0.1",
		Port:    6380,
		Auth:    "",
		Timeout: 3,
	})

	defer c.Close()
	if err != nil {
		log.Printf("conn err:%s\n", err.Error())
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
