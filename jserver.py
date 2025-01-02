import sys
import socket
import logging
import threading
import json
import struct

HEAD_SIZE = 4
logging.basicConfig(encoding='utf-8', level=logging.DEBUG)


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


def parse_request(req):
    rfile = req.makefile("rb", -1)
    thread_id = threading.current_thread().ident
    counter = 0
    while True:
        h = rfile.read(HEAD_SIZE)
        if len(h) == 0:
            logging.info("eof received")
            break
        size = struct.unpack('>i', h)[0]
        body = rfile.read(size)
        bint = json.loads(body)
        logging.info("thread %s receive  %s", thread_id, bint)
        handle_request(req, bint)
        counter += 1
    logging.info("thread %s req %s times", thread_id, counter)
    req.close()


if __name__ == "__main__":
    soc = socket.socket(socket.AF_INET)
    soc.bind(("localhost", 8888))
    soc.listen(1024)
    logging.info("start listen at 8888")
    while True:
        req, addr = soc.accept()
        logging.info("accept from %s", addr)
        t = threading.Thread(target=parse_request, args=(req,))
        t.start()
