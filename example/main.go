package main

import (
	"log"
	"time"

	"github.com/lifezq/ssdbcl"
)

func main() {

	c, err := ssdbcl.New(ssdbcl.Config{
		Host:    "127.0.0.1",
		Port:    5069,
		Auth:    "b738fcf739f64ad86feb468045f3ce3e",
		Timeout: 3,
	})

	defer c.Close()

	if err != nil {
		log.Printf("conn err:%s\n", err.Error())
	}

	n, err := c.Get("mZ0asys/master/leader")

	log.Printf("n:%d err:%v\n", n, err)

	time.Sleep(2e9)
}
