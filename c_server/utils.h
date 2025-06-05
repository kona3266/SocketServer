#ifndef UTILS_H
#define UTILS_H

#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>

int listen_inet_socket(int port);

void report_peer_connected(const struct sockaddr_in* sa, socklen_t salen);
#endif