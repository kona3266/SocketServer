// thread socket server
#include <stdint.h>
#include <sys/socket.h>
#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <string.h>
#include <netinet/in.h>
#include <pthread.h>
#include <unistd.h>
#include <errno.h>

#include "utils.h"

typedef enum
{
    WAIT_FOR_MSG,
    IN_MSG
} InternalState;

typedef struct {int sockfd;} thread_config_t;

void reverse(char *str, int len) {
    for (int i = 0, j = len - 1; i < j; i++, j--) {
        char tmp = str[i];
        str[i] = str[j];
        str[j] = tmp;
    }
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

ssize_t send_all(int sockfd, void *buf, size_t len) {
    size_t total_sent = 0;
    char *ptr = (char *)buf;

    while (total_sent < len) {
        ssize_t sent = send(sockfd, ptr + total_sent, len - total_sent, 0);
        if (sent == -1) {
            if (errno == EINTR) continue;  // interrupt by signal, try again
            printf("send error\n");
            return total_sent;
        }
        total_sent += sent;
    }
    return total_sent;
}

void handle_connection(int sockfd)
{
    uint8_t head[4];
    uint32_t net_num;
    int body_len=0;
    int rep_len=0;
    int offset = 0;
    uint32_t len = 0;
    int val = 0;
    while (1)
    {
        // read head first

        while(offset < 4) {
            // read until head is full
            len = recv(sockfd, head+offset, 4-offset, 0);
            offset += len;
            if (len < 0)
            {
                perror("recv");
                exit(1);
            }else if (len == 0)
            {
                //printf("EOF exit\n");
                if (errno == EINTR) continue;
                goto close_connect;
            }
            if (offset == 4){
                offset = 0;
                break;
            }
        }
        memcpy(&net_num, head, 4);
        body_len = (int)ntohl(net_num);
        char *body = calloc(body_len+1, 1);
        body[body_len] = '\0';
        // read body
        while(offset < body_len) {
            len = recv(sockfd, body+offset, body_len-offset, 0);
            offset += len;
            if (len < 0)
            {
                perror("recv");
                exit(1);
            }
            else if (len == 0)
            {
                if (errno == EINTR) continue;
                printf("error offset %d, body len %d\n", offset, body_len);
                fflush(stdout);
                goto close_connect;
            }
            if (offset == body_len){
                offset = 0;
                break;
            }
        }
        val = atoi(body);
        free(body);
        // return val + 1 to client
        rep_len = count_digits(val+1);

        char *rep = calloc(rep_len, 1);
        itoa(val+1, rep);
        len = htonl(rep_len);
        if (send_all(sockfd, &len, sizeof(len)) != sizeof(len)) {
            printf("send head error\n");
            fflush(stdout);
            goto close_connect;
        }
        if (send_all(sockfd, rep, rep_len) != rep_len) {
            printf("send body error\n");
            fflush(stdout);
            goto close_connect;
        };
        free(rep);
    }
close_connect:
    close(sockfd);
}


void *server_thread(void *arg) {
    thread_config_t* config = (thread_config_t*) arg;
    int sockfd = config->sockfd;
    free(config);
    unsigned long id = (unsigned long) pthread_self();
    //printf("Thread %lu created to handle connection with socket %d\n", id, sockfd);
    handle_connection(sockfd);
    // printf("Thread %ld done\n", id);
    // fflush(stdout); 
    return NULL;
}

void sig_handler(int sig){
    if (sig == SIGINT) {
        printf("received SIGINT exit\n");
        exit(0);
    }
}

int main(int argc, char **argv)
{
    int port = 9090;
    if (argc >= 2)
    {
        port = atoi(argv[1]);
    }
    if (signal(SIGINT, sig_handler) == SIG_ERR){
        printf("can't catch SIGINT\n");
    }
    int sockfd = listen_inet_socket(port);
    while (1)
    {
        struct sockaddr_in peer_addr;
        socklen_t peer_addr_len = sizeof(peer_addr);

        // printf("waiting for connect...\n");

        int newsockfd = accept(sockfd, (struct sockaddr *)&peer_addr, &peer_addr_len);
        if (newsockfd < 0)
        {
            perror("ERROR on accept");
            exit(1);
        }

        //report_peer_connected(&peer_addr, peer_addr_len);
        pthread_t the_thread;
        thread_config_t* config = (thread_config_t*)malloc(sizeof(*config));

        if (!config) {
            perror("OOM");
            exit(1);
        }
        config->sockfd = newsockfd;
        if (pthread_create(&the_thread, NULL, server_thread, config) != 0){
            perror("pthread create");
            close(newsockfd);
            free(config);
        }else if (pthread_detach(the_thread) != 0){
            perror("pthread detach");
            close(newsockfd);
            free(config);

        }

    }
    return 0;
}
