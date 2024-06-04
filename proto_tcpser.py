from socket import *
import target_pb2

HOST = 'localhost'
PORT = 7777
BUFSIZ = 1024
ADDR = (HOST, PORT)
tcpSerSock = socket(AF_INET, SOCK_STREAM)
tcpSerSock.bind(ADDR)
tcpSerSock.listen(1000) #listen的参数代表，在连接被拒绝或者转发之前，传入连接的最大数
while True:
    print('waiting for connnection ...')
    tcpCliSock, addr = tcpSerSock.accept()
    print('... connected from: ', addr)
    while True:
        try:
            message = tcpCliSock.recv(BUFSIZ)
            if len(message) == 0:
                break
            recv_rec = target_pb2.Rect()
            recv_rec.ParseFromString(message)
            print(recv_rec.x1, recv_rec.y1)
            print(recv_rec.x2, recv_rec.y2)
            tcpCliSock.send(b"new message")
            
        except Exception as e:
            break
    print('connection closed from:', addr)
    tcpCliSock.close()