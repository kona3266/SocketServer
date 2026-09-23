#include <stdio.h>
#include <fcntl.h>
#include <assert.h>
#include <errno.h>
#include <stdlib.h>
#include <stdbool.h>
#include <sys/epoll.h>
#include <string.h>
#include <unistd.h>
#include "utils.h"


#define MAXFDS 40 * 1000


typedef enum
{
    READ_HEAD,
    READ_BODY,
    WRITE_HEAD,
    WRITE_BODY
} State;

typedef struct
{
    bool want_read;
    bool want_write;
} fd_status_t;

typedef struct {
    int fd;
    uint8_t head[4];
    char read_buf[1024];
    char write_buf[1024];
    int body_size;
    int offset;
    State cur_state;
} connection_t;

const fd_status_t fd_status_R = {.want_read = true, .want_write = false};
const fd_status_t fd_status_W = {.want_read = false, .want_write = true};
const fd_status_t fd_status_RW = {.want_read = true, .want_write = true};
const fd_status_t fd_status_NORW = {.want_read = false, .want_write = false};

void make_sockfd_non_blocking(int sockfd)
{
    int flags = fcntl(sockfd, F_GETFL, 0);
    if (flags == -1)
    {
        perror("fcntl F_GETFL");
        exit(1);
    }
    if (fcntl(sockfd, F_SETFL, flags | O_NONBLOCK) == -1)
    {
        perror("fcntl F_SETFL O_NONBLOCK");
        exit(1);
    }
}

void reverse(char *str, int len) {
    for (int i = 0, j = len - 1; i < j; i++, j--) {
        char tmp = str[i];
        str[i] = str[j];
        str[j] = tmp;
    }
}

int count_digits(int num) {
    if (num == 0) return 1; 

    int count = 0;
    num = abs(num);

    while (num > 0) {
        num /= 10;
        count++;
    }
    return count;
}

void itoa(int num, char *str) {
    int i = 0;
    if (num == 0) {
        str[i] = '0';
        return;
    }

    while (num > 0) {
        str[i] = '0' + (num % 10);
        i++;
        num /= 10;
    }
    reverse(str, i);
}

fd_status_t handle_conn(connection_t *conn){
    switch (conn->cur_state)
    {
    case READ_HEAD:
        {
        if(conn->offset < 4) {
            int len = recv(conn->fd, conn->head + conn->offset, 4 - conn->offset, 0);
            if (len < 0)
            {
                if (errno == EAGAIN || errno == EWOULDBLOCK)
                {
                    return fd_status_R;
                } else {
                    perror("recv");
                    exit(1);
                }
            }else if (len == 0)
            {
                return fd_status_NORW;
            }
            conn->offset += len;
            if (conn->offset == 4){
                conn->offset = 0;
            } else {
                return fd_status_R;
            }
        }
        uint32_t net_num;
        memcpy(&net_num, conn->head, 4);
        conn->body_size = (int)ntohl(net_num);
        conn->read_buf[conn->body_size] = '\0';
        conn->cur_state = READ_BODY;
        return fd_status_R;
        }

    case READ_BODY:
        {
        if (conn->offset < conn->body_size){
            int len = recv(conn->fd, conn->read_buf + conn->offset, conn->body_size - conn->offset, 0);
            if (len < 0)
            {
                if (errno == EAGAIN || errno == EWOULDBLOCK)
                {
                    return fd_status_R;
                } else {
                    perror("recv");
                    exit(1);
                }
            }else if (len == 0)
            {
                return fd_status_NORW;
            }
            conn->offset += len;
            if (conn->offset == conn->body_size){
                conn->offset = 0;
            } else {
                return fd_status_R;
            }
        }
        // read val and then clear read buf;
        int val = atoi(conn->read_buf);
        memset(conn->read_buf, 0, sizeof(conn->read_buf));
        int rep_len = count_digits(val+1);
        conn->body_size = rep_len;
        uint32_t net_num;
        net_num = htonl((u_int32_t)rep_len);
        memcpy(conn->head, &net_num, 4);
        itoa(val+1, conn->write_buf);
        conn->cur_state = WRITE_HEAD;
        return fd_status_W;
    }

    case WRITE_HEAD:
    {
        if (conn->offset < 4){
            int len = send(conn->fd, conn->head + conn->offset, 4 - conn->offset, 0);
            if (len < 0)
            {
                if (errno == EAGAIN || errno == EWOULDBLOCK)
                {
                    return fd_status_W;
                } else {
                    perror("recv");
                    exit(1);
                }
            }else if (len == 0)
            {
                return fd_status_NORW;
            }
            conn->offset += len;
            if (conn->offset == 4){
                conn->offset = 0;
            } else {
                return fd_status_W;
            }
        }
        conn->cur_state = WRITE_BODY;
        return fd_status_W;
    }

    case WRITE_BODY:
    {
        if (conn->offset < conn->body_size){
            int len = send(conn->fd, conn->write_buf + conn->offset, conn->body_size - conn->offset, 0);
            if (len < 0)
            {
                if (errno == EAGAIN || errno == EWOULDBLOCK)
                {
                    return fd_status_W;
                } else {
                    perror("recv");
                    exit(1);
                }
            }else if (len == 0)
            {
                return fd_status_NORW;
            }
            conn->offset += len;
            if (conn->offset == conn->body_size){
                conn->offset = 0;
            } else {
                return fd_status_W;
            }
        }
        conn->body_size = 0;
        conn->cur_state = READ_HEAD;
        return fd_status_R;
    }
    }
}


int main(int argc, char **argv){
    setvbuf(stdout, NULL, _IONBF, 0);
    int portnum = 9090;
    if (argc >= 2)
    {
        portnum = atoi(argv[1]);
    }
    printf("serve on port %d\n", portnum);

    int listener_fd = listen_inet_socket(portnum);
    make_sockfd_non_blocking(listener_fd);
    int epollfd = epoll_create1(0);

    if (epollfd < 0)
    {
        perror("epoll_create1");
        exit(1);
    }
    struct epoll_event accept_event;
    connection_t *listen_conn = malloc(sizeof(connection_t));
    listen_conn->fd = listener_fd;
    accept_event.data.ptr = listen_conn;
    accept_event.events = EPOLLIN;
    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, listener_fd, &accept_event) < 0)
    {
        perror("epollctl EPOLL_CTL_ADD");
        exit(1);
    }
    struct epoll_event *events = calloc(MAXFDS, sizeof(struct epoll_event));
    if (events == NULL)
    {
        printf("unable to allocate mem for epoll_events");
        exit(1);
    }

    while (1)
    {
        int nready = epoll_wait(epollfd, events, MAXFDS, -1);
        for (int i = 0; i < nready > 0; i++)
        {
            if (events[i].events & EPOLLERR)
            {
                perror("epoll_wait EPOLLERR");
            }
            connection_t *conn = (connection_t *)events[i].data.ptr;

            if (conn->fd == listener_fd)
            {
                struct sockaddr_in peer_addr;
                socklen_t peer_addr_len = sizeof(peer_addr);
                int newsockfd = accept(listener_fd, (struct sockaddr *)&peer_addr, &peer_addr_len);

                if (newsockfd < 0)
                {
                    if (errno == EAGAIN || errno == EWOULDBLOCK)
                    {
                        printf("accept returned EAGAIN or EWOULDBLOCK\n");
                    }
                    else
                    {
                        perror("accept");
                        exit(1);
                    }
                }else{
                    make_sockfd_non_blocking(newsockfd);
                    if (newsockfd > MAXFDS)
                    {
                        printf("socket fd %d > MAXFDS", newsockfd);
                        exit(1);
                    }
                    connection_t *new_conn = malloc(sizeof(connection_t));
                    new_conn->fd = newsockfd;
                    new_conn->cur_state = READ_HEAD;
                    memset(new_conn->read_buf, 0, sizeof(new_conn->read_buf));
                    struct epoll_event event = {0};
                    event.data.ptr = new_conn;
                    event.events |= EPOLLIN;

                    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, newsockfd, &event) < 0)
                    {
                        perror("epoll_ctl EPOLL_CTL_ADD");
                        exit(1);
                    }
                }
            } else {
                if (events[i].events & EPOLLIN) {
                    fd_status_t status = handle_conn(conn);
                    struct epoll_event event = {0};
                    event.data.ptr = conn;
                    if (status.want_read) {
                        event.events |= EPOLLIN;
                    }
                    if (status.want_write) {
                        event.events |= EPOLLOUT;
                    }

                    if (event.events == 0)
                    {
                        // printf("socket %d closing\n", conn->fd);
                        if (epoll_ctl(epollfd, EPOLL_CTL_DEL, conn->fd, NULL) < 0)
                        {
                            printf("%d ", conn->fd);
                            perror("epoll_ctl EPOLL_CTL_DEL");
                            exit(1);
                        }
                        close(conn->fd);
                        free(conn);
                    }
                    else if (epoll_ctl(epollfd, EPOLL_CTL_MOD, conn->fd, &event) < 0)
                    {
                        printf("%d ", conn->fd);
                        perror("epoll_ctl EPOLL_CTL_MOD");
                        exit(1);
                    }
                }else if (events[i].events & EPOLLOUT){
                    fd_status_t status = handle_conn(conn);
                    struct epoll_event event = {0};
                    event.data.ptr = conn;

                    if (status.want_read) {
                        event.events |= EPOLLIN;
                    }
                    if (status.want_write) {
                        event.events |= EPOLLOUT;
                    }


                    if (event.events == 0)
                    {
                        printf("socket %d closing\n", conn->fd);
                        if (epoll_ctl(epollfd, EPOLL_CTL_DEL, conn->fd, NULL) < 0)
                        {
                            printf("%d ", conn->fd);
                            perror("epoll_ctl EPOLL_CTL_DEL");
                            exit(1);
                        }
                        close(conn->fd);
                        free(conn);
                    }
                    else if (epoll_ctl(epollfd, EPOLL_CTL_MOD, conn->fd, &event) < 0)
                    {
                        printf("%d ", conn->fd);
                        perror("epoll_ctl EPOLL_CTL_MOD");
                        exit(1);
                    }
                }
            }
        }
    }
    free(listen_conn);
    free(events);
    return 0;

}