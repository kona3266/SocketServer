package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net"
	"sync"
)

type Conn struct {
	tconn io.ReadWriteCloser
	buf   io.Writer
	rd    *bufio.Reader
}

type Writer struct {
	buf io.Writer
}

type Reader struct {
	rd *bufio.Reader
}

func CheckEof(data []byte) bool {
	if data[0] == byte('-') {
		return true
	}
	return false
}

func Read2byte(rd *bufio.Reader, data []byte) error {
	i := 0
	for {
		n, err := rd.Read(data[i:])
		i += n
		if err != nil {
			return err
		}
		if i == 2 {
			break
		}
	}
	return nil

}
func (r *Reader) Read(data []byte) (int, error) {

	// 自定义read，检查EOF标志
	length := len(data)
	merged_data := make([]byte, 2)
	i := 0
	for {
		// read buf first, then io reader
		err := Read2byte(r.rd, merged_data)

		if err != nil {
			fmt.Println(err)
			return i, err
		}
		if CheckEof(merged_data) {
			fmt.Println("read", i, "buf size", r.rd.Buffered())
			return i, io.EOF
		}
		copy(data[i:i+1], []byte{merged_data[1]})
		i += 1
		if i == length {
			return length, nil
		}
		merged_data = make([]byte, 2)
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {

	var b bytes.Buffer
	for _, v := range p {
		b.Write([]byte{48, v})
	}

	for {
		nl, err := b.WriteTo(w.buf)
		if err == nil {
			return len(p), nil
		}
		if err != io.ErrShortWrite {
			return int(nl), err
		}
	}
}

func (w *Writer) Close() error {

	// 插入eof 结束符
	eof := "-"
	w.buf.Write([]byte(eof))
	w.buf.Write([]byte(eof))
	return nil
}

func (conn *Conn) Send(key string) (writer io.WriteCloser, err error) {
	writer = &Writer{buf: conn.buf}
	_, err = writer.Write([]byte(key))
	if err != nil {
		return nil, err
	}
	writer.Close()
	return writer, nil
}

func (conn *Conn) Receive() (key string, reader io.Reader, err error) {
	myreader := &Reader{rd: conn.rd}

	data := make([]byte, 512)

	n, err := myreader.Read(data)
	if err == io.EOF && n != 0 {
		err = nil
	}

	if err != nil {
		fmt.Println("read msg failed, err:", err)
		return "", nil, err
	}

	return string(data[:n]), myreader, err
}

func (conn *Conn) Close() {
	conn.tconn.Close()
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{tconn: conn,
		buf: conn,
		rd:  bufio.NewReader(conn),
	}
}

func dial(serverAddr string) *Conn {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	return NewConn(conn)
}

func startServer(handle func(*Conn)) net.Listener {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("[WARNING] ln.Accept", err)
				return
			}
			go handle(NewConn(conn))
		}
	}()
	return ln
}

func assertEqual(actual string, expected string) {
	if actual != expected {
		panic(fmt.Sprintf("actual:%v expected:%v\n", actual, expected))
	}
}

func newRandomKey() string {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func readRandomData(reader io.Reader, hash hash.Hash) (checksum string) {
	hash.Reset()
	var buf = make([]byte, 23<<20)
	for {
		n, err := reader.Read(buf)
		fmt.Println("read", n, "max read size", len(buf), err)

		if err == io.EOF {
			_, err = hash.Write(buf[:n])
			if err != nil {
				panic(err)
			}
			break
		}
		if err != nil {
			panic(err)
		}

		_, err = hash.Write(buf[:n])
		if err != nil {
			panic(err)
		}
	}
	checksum = hex.EncodeToString(hash.Sum(nil))
	return checksum
}

func writeRandomData(writer io.Writer, hash hash.Hash) (checksum string) {
	hash.Reset()
	const (
		dataSize = 500 << 20
		bufSize  = 1 << 20
	)
	var (
		buf  = make([]byte, bufSize)
		size = 0
	)
	for i := 0; i < dataSize/bufSize; i++ {
		_, err := rand.Read(buf)
		if err != nil {
			panic(err)
		}
		_, err = hash.Write(buf)
		if err != nil {
			panic(err)
		}
		n, err := writer.Write(buf)
		if err != nil {
			panic(err)
		}
		size += n
	}
	if size != dataSize {
		panic(size)
	}
	fmt.Println("write size", size)
	checksum = hex.EncodeToString(hash.Sum(nil))
	return checksum
}

func testCase1() {
	var (
		mapKeyToChecksum = map[string]string{}
		lock             sync.Mutex
	)
	ln := startServer(func(conn *Conn) {
		key, reader, err := conn.Receive()
		if err != nil {
			panic(err)
		}
		var (
			h         = sha256.New()
			_checksum = readRandomData(reader, h)
		)
		fmt.Println("check key in server", _checksum)
		lock.Lock()
		checksum, keyExist := mapKeyToChecksum[key]
		lock.Unlock()
		if !keyExist {
			panic(fmt.Sprintln(key, "not exist"))
		}
		assertEqual(_checksum, checksum)

		for _, key := range []string{newRandomKey(), newRandomKey()} {
			writer, err := conn.Send(key)
			if err != nil {
				panic(err)
			}
			checksum := writeRandomData(writer, h)
			lock.Lock()
			mapKeyToChecksum[key] = checksum
			lock.Unlock()
			err = writer.Close()
			if err != nil {
				panic(err)
			}
		}
		conn.Close()
	})
	//goland:noinspection GoUnhandledErrorResult
	defer ln.Close()

	conn := dial(ln.Addr().String())

	var (
		key = newRandomKey()
		h   = sha256.New()
	)

	writer, err := conn.Send(key)
	if err != nil {
		panic(err)
	}
	checksum := writeRandomData(writer, h)
	fmt.Println("client wirte ..", checksum)
	lock.Lock()
	mapKeyToChecksum[key] = checksum
	lock.Unlock()
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	for {
		key, reader, err := conn.Receive()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		_checksum := readRandomData(reader, h)
		lock.Lock()
		checksum, keyExist := mapKeyToChecksum[key]
		lock.Unlock()
		if !keyExist {
			panic(fmt.Sprintln(key, "not exist"))
		}
		assertEqual(_checksum, checksum)
	}
	fmt.Println("test 1 finished")
	conn.Close()
}

func main() {
	testCase1()
}
