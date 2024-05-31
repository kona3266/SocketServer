package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

type tconn struct {
	conn io.ReadWriteCloser
	rd   *bufio.Reader
	wt   *bufio.Writer
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

func main() {
	rtcpaddr := os.Args[1]

	tcpaddr, err := net.ResolveTCPAddr("tcp4", rtcpaddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	c := tconn{conn: conn, rd: bufio.NewReader(conn), wt: bufio.NewWriter(conn)}
	c.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	result, _ := c.ReadMessage()
	fmt.Println(string(result))
	c.Close()

}

