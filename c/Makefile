CC      ?= gcc
CFLAGS  ?= -O2 -Wall
LDFLAGS  = -lpthread

BINS = thread_server epoll_server epoll_threadpool

all: $(BINS)

thread_server: thread_server.c utils.c utils.h
	$(CC) $(CFLAGS) -o $@ thread_server.c utils.c $(LDFLAGS)

epoll_server: epoll_server.c utils.c utils.h
	$(CC) $(CFLAGS) -o $@ epoll_server.c utils.c

epoll_threadpool: epoll_threadpool.c utils.c utils.h
	$(CC) $(CFLAGS) -o $@ epoll_threadpool.c utils.c $(LDFLAGS)

clean:
	rm -f $(BINS)

.PHONY: all clean
