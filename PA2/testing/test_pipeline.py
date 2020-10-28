from socket import socket

# Create connection to the server
s = socket()
s.connect(("localhost", 8080))

# Compose tests
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

# Test 5-second timeout for incomplete request after complete
s = socket()
s.connect(("localhost", 8080))
msgTimeout = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\nGET / HTTP/1."

# Send out the request
s.sendall(msgTimeout)

# Listen for response and print it out
print (s.recv(4096))

s.close()