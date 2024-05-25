package main

import (
        "fmt"
        "io"
        "net"
        "os"
)

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
        conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
        result, _ := io.ReadAll(conn)
        fmt.Println(string(result))

}
