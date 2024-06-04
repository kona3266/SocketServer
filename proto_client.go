package main

import (
	"fmt"
	pro "gobuf/proto"
	"net"

	"github.com/golang/protobuf/proto"
)

func main() {
	rect := &pro.Rect{
		X1: 1,
		X2: 2,
		Y1: 3,
		Y2: 4,
	}
	data, err := proto.Marshal(rect)
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", ":7777")
	if err != nil {
		panic(err)
	}
	conn.Write(data)
	rep := make([]byte, 30)
	conn.Read(rep)
	fmt.Println(string(rep))
	conn.Close()

}
