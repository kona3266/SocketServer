from soket import *
from time import ctime
HOST = ''
PORT = 21567
BUFSIZ = 1024
ADDR = (HOST, PORT)
tcpSerSock = socket(AF_INET, SOCK_STREAM)
tcpSerSock.bind(ADDR)
tcpSerSock.listen(5) #listen的参数代表，在连接被拒绝或者转发之前，传入连接的最大数
while True:
    print('waiting for connnection ...')
    tcpCliSock, addr = tcpSerSock.accept()
    print('... connected from: ', addr)
    while True:
        data = tcpCliSock.recv(BUFSIZ)
        if not data:
            break
        tcpCliSock.send('[%s] %s' % (bytes(ctime(), 'utf-8'), data))
    tcpCliSock.close()
tcpSerSock.close()
