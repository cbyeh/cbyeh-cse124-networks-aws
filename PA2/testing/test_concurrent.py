from socket import socket

# Create connection to the server
s1 = socket()
s2 = socket()
s1.connect(("localhost", 8080))
s2.connect(("localhost", 8080))

# Compose test for root
msg = b"GET / HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s1.sendall(msg)
s2.sendall(msg)

# Listen for response and print it out
print(s1.recv(4096))
print(s2.recv(4096))

s1.close()
s2.close()
