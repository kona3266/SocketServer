# coding=utf-8
import socket
import logging
import threading
import json
import struct

HEAD_SIZE = 4
logging.basicConfig(filename='./server_log', encoding='utf-8', level=logging.INFO)


''' packet format
             4bytes        length
           ┌────────┬─────────────────────┐
           │ length │      body           │
           └────────┴─────────────────────┘
'''

def handle_request(req, bint):
    wfile = req.makefile("wb", 0)
    # add 1 to the input, then write it back
    bint += 1
    bstr = json.dumps(bint)
    size = len(bstr)
    sizebytes = struct.pack('>i', size)
    wfile.write(sizebytes)
    wfile.write(bstr)


def parse_request(req, addr):
    rfile = req.makefile("rb", -1)
    thread_id = threading.current_thread().ident
    counter = 0
    while True:
        h = rfile.read(HEAD_SIZE)
        if len(h) == 0:
            logging.debug("eof received %s", addr)
            break
        size = struct.unpack('>i', h)[0]
        body = rfile.read(size)
        bint = json.loads(body)
        logging.debug("thread %s receive  %s", thread_id, bint)
        handle_request(req, bint)
        counter += 1
    if counter != 100:
        logging.info("addr %s req %s times", addr, counter)
    req.close()


if __name__ == "__main__":
    soc = socket.socket(socket.AF_INET)
    soc.bind(("localhost", 8888))
    soc.listen(50000)
    rcvbuf_size = soc.getsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF)
    
    logging.info("start listen at 8888, rcv buffer [%s]KB", rcvbuf_size/1024)
    count = 0
    while True:
        logging.debug("handled [%s] req", count)
        req, addr = soc.accept()
        logging.debug("accept from %s", addr)
        count += 1
        t = threading.Thread(target=parse_request, args=(req, addr), name=addr)
        t.start()



