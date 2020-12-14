from socket import *
from time import ctime
import threading

HOST = ''
PORT = 21567
BUFSIZ = 1024
ADDR = (HOST, PORT)
udpSerSock = socket(AF_INET, SOCK_DGRAM)
udpSerSock.bind(ADDR)
def recv_handle(udpSerSock, BUFSIZ,addr):
    while True:
        data, addr = udpSerSock.recvfrom(BUFSIZ)
        if data:
            data = data.decode()
            print(addr,'says:',data)
        reply = input('>')
        udpSerSock.sendto(reply.encode(), addr)
        break
    print('... close from ', addr)

#udp是无连接的，没有独立的套接字通道，也不需要转换操作。
#因此，udp发送消息时，也需要指定发送的地址。

while True:
    print('waiting for message....')
    data, addr = udpSerSock.recvfrom(BUFSIZ)
    threading.Thread(target=recv_handle,args=(udpSerSock, BUFSIZ, addr)).start()

    

