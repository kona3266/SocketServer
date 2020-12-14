from socket import *
from time import ctime
import threading

HOST = 'localhost'
PORT = 21567
BUFSIZ = 1024
ADDR = (HOST, PORT)
udpCliSock = socket(AF_INET, SOCK_DGRAM)
udpCliSock.bind(('localhost',21568))
def recv(udpCliSock):
    while True:
        data, ADDR = udpCliSock.recvfrom(BUFSIZ)
        if data:
            print(ADDR, 'says:', data.decode())
            print('>')
def send(udpCliSock,ADDR):
    while True:
        msg = input('>')
        if not msg:
            break
        udpCliSock.sendto(msg.encode(), ADDR)
udpCliSock.sendto('initializing'.encode(), ADDR)
threading.Thread(target=send, args=(udpCliSock, ADDR)).start()
threading.Thread(target=recv, args=(udpCliSock,)).start()
