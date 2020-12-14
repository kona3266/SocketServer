import cmdFunc
import http.server
from socket import *
from time import ctime
from threading import Thread

def requesthandler(tcpClisock, addr):
    print('... connected from: ', addr)
    while True:
        try:
            data = tcpCliSock.recv(BUFSIZ)
            data = data.decode()
            data = cmdFunc.cmdExcute(data)
            if not data:
                break
            tcpCliSock.send(('[%s] %s' % (ctime(), data)).encode())
        except Exception as e:
            break
    print('connection closed from:', addr)
    tcpCliSock.close()

HOST = 'localhost'
PORT = 21567
BUFSIZ = 1024
ADDR = (HOST, PORT)
tcpSerSock = socket(AF_INET, SOCK_STREAM)
tcpSerSock.bind(ADDR)
tcpSerSock.listen(1000) #listen的参数代表，在连接被拒绝或者转发之前，传入连接的最大数
while True:
    print('waiting for connnection ...')
    tcpCliSock, addr = tcpSerSock.accept()
    t=Thread(target=requesthandler, args=(tcpCliSock, addr))
    t.start()

