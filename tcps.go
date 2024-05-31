package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

type tconn struct {
	conn io.ReadWriteCloser
	rd   *bufio.Reader
	wt   *bufio.Writer
}

func main() {
	service := ":7777"
	tcpaddr, _ := net.ResolveTCPAddr("tcp4", service)

	listener, _ := net.ListenTCP("tcp", tcpaddr)
	for {
		conn, _ := listener.Accept()
		c := tconn{conn: conn, rd: bufio.NewReader(conn), wt: bufio.NewWriter(conn)}
		go handleClient(c)
	}

}

func (c tconn) Close() error {
	return c.conn.Close()
}

func (c tconn) ReadMessage() ([]byte, error) {
	msg, err := c.rd.ReadBytes(byte('*'))
	if err == nil {
		msg = msg[:len(msg)-1]
	}

	return msg, err
}

func (c tconn) Write(data []byte) (int, error) {
	n, err := c.wt.Write(data)
	c.wt.WriteByte(byte('*'))
	c.wt.Flush()
	return n, err
}

func handleClient(conn tconn) {
	// close tcp conn after handle
	defer conn.Close()
	for {

		request, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read err", err)
			break
		}

		fmt.Println(string(request))
		daytime := time.Now().String()
		fmt.Println(daytime)
		conn.Write([]byte(daytime))
	}

}
