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

func NewConn(conn net.Conn) tconn {
	return tconn{conn: conn, rd: bufio.NewReader(conn), wt: bufio.NewWriter(conn)}
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

func handleClient(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("ln.Accept", err)
			return
		}
		go handleConn(NewConn(conn))
	}
}

func handleConn(conn tconn) {
	defer conn.Close()
	for {
		request, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read err", err)
			break
		}

		fmt.Println("server read", string(request))
		daytime := time.Now().String()
		conn.Write([]byte(daytime))
	}
}

func startServer() net.Listener {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go handleClient(ln)
	return ln
}

func main() {
	ln := startServer()
	defer ln.Close()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		panic(err)
	}

	c := NewConn(conn)
	c.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	result, _ := c.ReadMessage()
	fmt.Println("client read", string(result))
	c.Close()

}
