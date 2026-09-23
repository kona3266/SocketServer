[![License](https://img.shields.io/badge/license-Apache%202-green.svg)](https://www.apache.org/licenses/LICENSE-2.0)

# SocketServer

一个用 **Go / C / Python** 三种语言实现的 TCP 长度前缀（length-prefixed）服务器集合，用于对比不同语言、不同并发模型下的 Socket 服务端写法与性能。

所有服务端实现同一套业务逻辑：**接收一个整数，将其加 1 后原路返回**。

## 协议

每个报文由 4 字节大端序（big-endian）长度头 + 消息体组成：

```
             4 bytes             N bytes
           ┌────────┬───────────────────────┐
           │ length │         body          │
           └────────┴───────────────────────┘
```

- 消息体为整数的文本表示（如 `42`、`-7`）。Go/Python 端按 JSON 编解码，与纯数字文本在线路上兼容；
- 服务端在**同一条连接上**循环处理多个报文（支持连接复用）；
- 客户端（`go/jclient.go`）在每次请求后校验响应值是否为请求值 + 1。

## 目录结构

```
SocketServer/
├── go/                     # Go 实现（goroutine-per-connection）
│   ├── go.mod
│   ├── jserver.go          # TCP 服务器：监听 127.0.0.1:8888，每连接一个 goroutine
│   ├── jclient.go          # 压测客户端：单连接发送 100 个随机数并校验响应
│   └── jclient_test.go     # 压力测试：并发启动 10000 个客户端连接
├── go-tlv/                 # Go 实验代码：通用 LV（Length-Value）流式读写封装
│   ├── go.mod
│   └── raw_tcp_LV.go       # Conn/Reader/Writer 抽象，支持大数据（500MB）传输校验
├── c/                      # C 实现（三种并发模型）
│   ├── Makefile
│   ├── utils.c / utils.h   # 公共工具：创建监听 socket 等
│   ├── thread_server.c     # 模型一：thread-per-connection（默认端口 9090）
│   ├── epoll_server.c      # 模型二：单线程 epoll 事件循环（非阻塞 IO + 状态机）
│   └── epoll_threadpool.c  # 模型三：epoll + 固定线程池（4 worker，各自持有 epoll 实例）
├── python/                 # Python 实现
│   ├── jserver.py          # thread-per-connection（监听 localhost:8888，backlog 50000）
│   └── async_jserver.py    # asyncio 单线程事件循环实现
├── LICENSE
└── README.md
```

> 历史目录映射：`c_server/` → `c/`，`python_server/` → `python/`，`tcp_TLV/` → `go-tlv/`；
> 原 `c_server/jserver.c`（thread-per-connection 实现）更名为 `c/thread_server.c`，
> 原 `c_server/thread_epoll.c` 更名为 `c/epoll_threadpool.c`。

## 构建与运行

### Go

要求 Go 1.22+（客户端使用了 `math/rand/v2`）。

```bash
# 启动服务器（默认监听 127.0.0.1:8888）
cd go && go run jserver.go

# 运行压测（10000 并发客户端，每客户端 100 次请求）
cd go && go test -v

# 运行 LV 流式传输实验
cd go-tlv && go run raw_tcp_LV.go
```

### C

```bash
cd c && make          # 生成 thread_server / epoll_server / epoll_threadpool

./thread_server 9090  # thread-per-connection，端口通过 argv[1] 指定
./epoll_server 9090   # 单线程 epoll
./epoll_threadpool 9090
```

### Python

```bash
python3 python/jserver.py          # 多线程版（日志写入 ./server_log）
python3 python/async_jserver.py    # asyncio 版
```

### 快速验证

任一服务器启动后，可用 Python 一行式客户端验证协议：

```python
import socket, struct, json
s = socket.create_connection(("127.0.0.1", 8888))
b = json.dumps(42).encode()
s.sendall(struct.pack(">i", len(b)) + b)          # 发送 42
head = s.recv(4)
print(json.loads(s.recv(struct.unpack(">i", head)[0])))   # 输出 43
```

## 各实现对比

| 实现 | 并发模型 | 默认端口 | 备注 |
|------|----------|----------|------|
| `go/jserver.go` | goroutine-per-connection | 8888 | 代码最简洁 |
| `c/thread_server.c` | thread-per-connection | 9090 | 原型对照 |
| `c/epoll_server.c` | 单线程 epoll + 状态机 | 9090 | READ_HEAD → READ_BODY → WRITE_HEAD → WRITE_BODY 四状态 |
| `c/epoll_threadpool.c` | epoll + 4 线程池 | 9090 | 连接轮询分发到各 worker 的 epoll 实例 |
| `python/jserver.py` | thread-per-connection | 8888 | |
| `python/async_jserver.py` | asyncio 事件循环 | 8888 | 单线程高并发 |

## License

[Apache License 2.0](LICENSE)
