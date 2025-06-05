#include "utils.h"
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <string.h>


#define N_BACKLOG 50000

void report_peer_connected(const struct sockaddr_in* sa, socklen_t salen){
    int port = ntohs(sa->sin_port);
    char host[INET_ADDRSTRLEN];
    inet_ntop(AF_INET, &(sa->sin_addr), host, sizeof(host));
    printf("peer %s port %d connected\n", host, port);
}

int listen_inet_socket(int port) {
    int sockfd = socket(AF_INET, SOCK_STREAM, IPPROTO_IP);
    if (sockfd < 0) {
        perror("error open socket");
        exit(1);
    }
    // int opt = 1;
    // if (setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0){
    //     perror("setsockopt");
    //     exit(1);
    // }
    struct sockaddr_in serv_addr;
    memset(&serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_addr.s_addr = INADDR_ANY;
    serv_addr.sin_port = htons(port);

    if (bind(sockfd, (struct sockaddr*)&serv_addr, sizeof(serv_addr)) < 0) {
        perror("binding error");
        exit(1);
    }

    if (listen(sockfd, N_BACKLOG) < 0){
        perror("listen error");
        exit(1);
    }
    return sockfd;
}
