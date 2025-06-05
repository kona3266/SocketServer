# coding=utf-8
import logging
import json
import struct
import asyncio

HEAD_SIZE = 4
logging.basicConfig(filename='./server_log', encoding='utf-8', level=logging.DEBUG)


''' packet format
             4bytes        length
           ┌────────┬─────────────────────┐
           │ length │      body           │
           └────────┴─────────────────────┘
'''

async def handle_request(writer, bint):

    # add 1 to the input, then write it back
    bint += 1
    bstr = json.dumps(bint).encode()
    size = len(bstr)
    sizebytes = struct.pack('>i', size)
    writer.write(sizebytes)
    writer.write(bstr)
    await writer.drain()



async def parse_request(reader, writer):
    while True:
        h = await reader.read(HEAD_SIZE)
        if len(h) == 0:
            addr = writer.get_extra_info('peername')
            logging.debug("eof received %s", addr)
            break

        size = struct.unpack('>i', h)[0]
        body = await reader.read(size)
        bint = json.loads(body)
        await handle_request(writer, bint)
    writer.close()
    await writer.wait_closed()

async def main():
    server = await asyncio.start_server(parse_request, '127.0.0.1', 8888, backlog=50000)

    addrs = ', '.join(str(sock.getsockname()) for sock in server.sockets)
    logging.info(f'Serving on {addrs}')

    async with server:
        await server.serve_forever()


if __name__ == "__main__":
    asyncio.run(main())

