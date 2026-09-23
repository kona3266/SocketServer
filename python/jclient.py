# coding=utf-8
"""Client for the length-prefixed increment server.

Connects to a server, sends `--count` random integers over a single
connection and verifies that each response equals request + 1.

Usage:
    python3 jclient.py [--host 127.0.0.1] [--port 8888] [--count 100]
"""
import argparse
import json
import random
import socket
import struct

HEAD_SIZE = 4


def recv_exact(sock: socket.socket, n: int) -> bytes:
    """Read exactly n bytes from the socket."""
    buf = b""
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("connection closed by peer")
        buf += chunk
    return buf


def request(sock: socket.socket, val: int) -> int:
    """Send one integer, receive and return the response."""
    body = json.dumps(val).encode()
    sock.sendall(struct.pack(">i", len(body)) + body)

    size = struct.unpack(">i", recv_exact(sock, HEAD_SIZE))[0]
    resp = json.loads(recv_exact(sock, size))
    return resp


def main() -> None:
    parser = argparse.ArgumentParser(description="length-prefixed increment client")
    parser.add_argument("--host", default="127.0.0.1", help="server host")
    parser.add_argument("--port", type=int, default=8888, help="server port")
    parser.add_argument("--count", type=int, default=100, help="number of requests")
    args = parser.parse_args()

    with socket.create_connection((args.host, args.port), timeout=5) as sock:
        print(f"connected to {args.host}:{args.port}, {args.count} requests")
        for i in range(args.count):
            val = random.randint(100000, 999999)
            got = request(sock, val)
            if got != val + 1:
                raise AssertionError(f"ERROR: sent {val}, got {got}")

    print(f"all {args.count} requests ok")


if __name__ == "__main__":
    main()
