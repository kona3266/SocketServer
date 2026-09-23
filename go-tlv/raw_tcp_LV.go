package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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
	buf  io.Writer
	load bytes.Buffer
}

type Reader struct {
	rd      *bufio.Reader
	bodylen uint32
	offset  uint32
}

func (r *Reader) Reset() {
	r.bodylen = 0
	r.offset = 0
}

func (r *Reader) ReadHead() (int, error) {
	head := make([]byte, 4)
	i := 0
	for {
		n, err := r.rd.Read(head[i:])
		i += n
		if i == 4 {
			break
		}
		if err != nil {
			return i, err
		}
	}
	length := BytesToInt(head)
	return length, nil
}

func (r *Reader) Read(data []byte) (int, error) {

	// if head is not exist, read head first else read body
	if r.bodylen == 0 {
		fmt.Println("read head ", r.bodylen)
		n, err := r.ReadHead()
		r.bodylen = uint32(n)
		fmt.Println("body len ", r.bodylen)
		if err != nil {
			return n, err
		}
	}
	// expect read length is len(data), actual body length is r.bodylen - r.offset
	length := len(data)
	if r.bodylen-r.offset > uint32(length) {
		i := 0
		for {
			n, err := r.rd.Read(data[i:])
			r.offset += uint32(n)
			i += n
			if i == length {
				return length, nil
			}
			if err != nil {
				return i, err
			}
		}
	} else {
		i := 0
		expectLen := r.bodylen - r.offset
		for {
			n, err := r.rd.Read(data[i:expectLen])
			i += n
			r.offset += uint32(n)
			if uint32(i) == expectLen {
				r.Reset()
				err = io.EOF
				return int(expectLen), err
			}
			if err != nil {
				return i, err
			}
		}
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {

	//write to buffer first, when close is called flush to io.reader

	n, err = w.load.Write(p)
	return n, err
}

func IntToBytes(n int) []byte {
	x := int32(n)
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(x))
	return head
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func (w *Writer) Close() error {

	// write load length in head
	len := w.load.Len()
	fmt.Println("data len ", len)
	// len occupy 4 bytes
	head := IntToBytes(len)
	_, err := w.buf.Write(head)
	if err != nil {
		panic(err)
	}

	for {
		_, err := w.load.WriteTo(w.buf)
		if err == io.ErrShortWrite {
			continue
		}
		w.load.Reset()
		return err
	}
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
	myreader := &Reader{rd: conn.rd, bodylen: 0, offset: 0}

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
