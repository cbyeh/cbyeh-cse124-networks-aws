from socket import socket

# Create connection to the server
s = socket()
s.connect(("localhost", 8080))

# Compose test for root
msgPart1 = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart1)

# Listen for response and print it out
print(s.recv(4096))

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose test for subdirectory
msgPart2 = b"GET /subdir1/index.html HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart2)

# Listen for response and print it out
print(s.recv(4096))

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose test for nested subdirectories
msgPart3 = b"GET /subdir1/subdir11/index.html HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart3)

# Listen for response and print it out
print(s.recv(4096))

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose test for subdirectory root
msgPart4 = b"GET /subdir1/ HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart4)

# Listen for response and print it out
print(s.recv(4096))

""""""

s = socket()
s.connect(("localhost", 8080))

# Compose test for mime type jpg
msgPart5 = b"GET /kitten.jpg HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart5)

# Listen for response and print it out
print(s.recv(4096))
