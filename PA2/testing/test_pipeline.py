from socket import socket
import time

# Create connection to the server
s = socket()
s.connect(("localhost", 8080))

# Compose tests for pipelined request, with a Connection: close header
msgPart1 = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\n"
msgPart2 = b"GET /kitten.jpg HTTP/1.1\r\nHost: Ha\r\n\r\n"
msgPart3 = b"GET /index.html HTTP/1.1\r\nHost: Ha\r\nConnection: close\r\n\r\n"
msgPart4 = b"GET /shouldNotRequest/ HTTP/1.1\r\nHost: Ha\r\n\r\n"
messages = [msgPart1, msgPart2, msgPart3, msgPart4]

pipeline = ""
for msg in messages:
    pipeline += msg

# Send out the request
s.sendall(pipeline)

# Listen for response and print it out
print(s.recv(4096))

s.close()

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose for two good consecutive requests
msgConsecutive = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\nGET /subdir1/index.html HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgConsecutive)

# Listen for response and print it out
print(s.recv(4096))

s.close()

""""""

s = socket()
s.connect(("localhost", 8080))

# Test connection close for incomplete request after a complete
msgIncomplete = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\nGET / HTTP/1."

# Send out the request
s.sendall(msgIncomplete)

# Listen for response and print it out
print(s.recv(4096))

s.close()

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose test for leaving root and coming back
msgTimeout = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\n"
msgAfterTimeout = b"GET /shouldNotRequest HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the requests
s.sendall(msgTimeout)
time.sleep(5)
s.sendall(msgAfterTimeout)

# Listen for response and print it out
print(s.recv(4096))

s.close()
