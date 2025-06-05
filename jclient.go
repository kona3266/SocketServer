package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
)

func newMessage(b []byte, conn net.Conn) {
	len := len(b)
	x := int32(len)
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(x))
	conn.Write(head)
	conn.Write(b)
}

func verifyRsp(val int, conn net.Conn, seq int) error {
	head := make([]byte, 4)
	i := 0
	var err error
	var n int
	for {
		n, err = conn.Read(head[i:])
		i += n
		if i == 4 {
			break
		}
		if err != nil {
			fmt.Println("head read error", err, i, seq)
			return err
		}
	}

	h := binary.BigEndian.Uint32(head)
	x := int32(h)

	body := make([]byte, x)

	i = 0
	for {
		n, err := conn.Read(body[i:])
		i += n
		if i == int(x) {
			break
		}
		if err != nil {
			fmt.Println("body read", err)
			break
		}
	}

	var rval int
	json.Unmarshal(body, &rval)

	if rval-1 != val {
		fmt.Println("ERROR", rval, val)
	}
	return nil
}

func JsonClient() {
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		log.Fatalln(err)
	}

	// multipart message
	for i := 0; i < 100; i++ {
		randomval := 100000 + rand.IntN(899999)
		valB, err := json.Marshal(randomval)
		if err != nil {
			log.Fatalln(err)
		}

		newMessage(valB, conn)
		err = verifyRsp(randomval, conn, i)
		if err != nil {
			break
		}
	}
	conn.Close()

}
