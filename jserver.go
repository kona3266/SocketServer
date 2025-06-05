package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

func readmsg(conn net.Conn) (int, error) {
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
			if !errors.Is(err, io.EOF) {
				fmt.Println("head read error", err, i)
			}

			return 0, err
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
	return rval, nil
}

func sendmsg(num int, conn net.Conn) error {
	b, err := json.Marshal(num)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	len := len(b)
	x := int32(len)
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(x))
	conn.Write(head)
	conn.Write(b)
	return nil
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			for {
				num, err := readmsg(c)
				if err != nil {
					if errors.Is(err, io.EOF) {
						// fmt.Println("close conn")
						break
					} else {
						log.Fatal(err)
					}
				}
				err = sendmsg(num+1, c)
				if err != nil {
					log.Fatal(err)
				}
			}
			c.Close()

		}(conn)
	}
}
