from socket import *
from time import ctime
import threading

HOST = ''
PORT = 21567
BUFSIZ = 1024
ADDR = (HOST, PORT)
udpSerSock = socket(AF_INET, SOCK_DGRAM)
udpSerSock.bind(ADDR)
def recv(udpSerSock, BUFSIZ):
    while True:
        data, addr = udpSerSock.recvfrom(BUFSIZ)
        if data:
            data = data.decode()
            print(addr,'says:',data)
print('waiting for message....')
#udp是无连接的，没有独立的套接字通道，也不需要转换操作。
#因此，udp发送消息时，也需要指定发送的地址。
def send(udpSerSock, addr):
    while True:
        reply = input('>')
        udpSerSock.sendto(reply.encode(), addr)
data, addr = udpSerSock.recvfrom(BUFSIZ)
threading.Thread(target=recv,args=(udpSerSock, BUFSIZ)).start()
threading.Thread(target=send,args=(udpSerSock,addr)).start()

    

