package main

import (
        "net"
        "fmt"
        "time"
)

func main() {
        service := ":7777"
        tcpaddr, _ := net.ResolveTCPAddr("tcp4", service)

        listener, _ := net.ListenTCP("tcp", tcpaddr)
        for {
                conn, _ := listener.Accept()
                go handleClient(conn)
        }

}

func handleClient(conn net.Conn) {
        // close tcp conn after handle
        defer conn.Close()
        for {
        request := make([]byte, 256)
        read_len, _ := conn.Read(request)
        if read_len == 0 {
            break
        }
        fmt.Println(string(request[:read_len]))
        daytime := time.Now().String()
        conn.Write([]byte(daytime))
        }

}
