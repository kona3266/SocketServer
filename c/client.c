// client: connect to a length-prefixed server, send random integers,
// verify that every response equals request + 1.
//
// usage: ./client [host] [port] [count]
#include <stdint.h>
#include <sys/socket.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <unistd.h>
#include <time.h>
#include <errno.h>

#define HEAD_SIZE 4
#define BODY_MAX 64

static int connect_to(const char *host, const char *port) {
    struct addrinfo hints, *res, *rp;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    int err = getaddrinfo(host, port, &hints, &res);
    if (err != 0) {
        fprintf(stderr, "getaddrinfo: %s\n", gai_strerror(err));
        exit(1);
    }

    int sockfd = -1;
    for (rp = res; rp != NULL; rp = rp->ai_next) {
        sockfd = socket(rp->ai_family, rp->ai_socktype, rp->ai_protocol);
        if (sockfd < 0)
            continue;
        if (connect(sockfd, rp->ai_addr, rp->ai_addrlen) == 0)
            break;
        close(sockfd);
        sockfd = -1;
    }
    freeaddrinfo(res);
    if (sockfd < 0) {
        perror("connect");
        exit(1);
    }
    return sockfd;
}

static int send_all(int sockfd, const void *buf, size_t len) {
    const char *ptr = (const char *)buf;
    size_t total = 0;
    while (total < len) {
        ssize_t n = send(sockfd, ptr + total, len - total, 0);
        if (n < 0) {
            if (errno == EINTR)
                continue;
            perror("send");
            return -1;
        }
        total += (size_t)n;
    }
    return 0;
}

// read exactly len bytes; return 0 on success, -1 on EOF/error
static int read_full(int sockfd, void *buf, size_t len) {
    char *ptr = (char *)buf;
    size_t total = 0;
    while (total < len) {
        ssize_t n = recv(sockfd, ptr + total, len - total, 0);
        if (n < 0) {
            if (errno == EINTR)
                continue;
            perror("recv");
            return -1;
        }
        if (n == 0)
            return -1; // EOF
        total += (size_t)n;
    }
    return 0;
}

// send one request, verify response == val + 1; return 0 on success
static int request(int sockfd, int val) {
    char body[BODY_MAX];
    int body_len = snprintf(body, sizeof(body), "%d", val);

    uint32_t net_len = htonl((uint32_t)body_len);
    if (send_all(sockfd, &net_len, HEAD_SIZE) != 0)
        return -1;
    if (send_all(sockfd, body, (size_t)body_len) != 0)
        return -1;

    uint8_t head[HEAD_SIZE];
    if (read_full(sockfd, head, HEAD_SIZE) != 0)
        return -1;
    uint32_t resp_len = ntohl(*(uint32_t *)head);
    if (resp_len >= BODY_MAX) {
        fprintf(stderr, "response too long: %u\n", resp_len);
        return -1;
    }

    char resp[BODY_MAX];
    if (read_full(sockfd, resp, resp_len) != 0)
        return -1;
    resp[resp_len] = '\0';

    int got = atoi(resp);
    if (got != val + 1) {
        fprintf(stderr, "ERROR: sent %d, got %d\n", val, got);
        return -1;
    }
    return 0;
}

int main(int argc, char **argv) {
    const char *host = argc >= 2 ? argv[1] : "127.0.0.1";
    const char *port = argc >= 3 ? argv[2] : "8888";
    int count = argc >= 4 ? atoi(argv[3]) : 100;

    srand((unsigned)time(NULL));
    int sockfd = connect_to(host, port);
    printf("connected to %s:%s, %d requests\n", host, port, count);

    for (int i = 0; i < count; i++) {
        int val = 100000 + rand() % 900000;
        if (request(sockfd, val) != 0) {
            fprintf(stderr, "request %d/%d failed\n", i + 1, count);
            close(sockfd);
            return 1;
        }
    }
    printf("all %d requests ok\n", count);
    close(sockfd);
    return 0;
}
