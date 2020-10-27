from socket import socket

# Create connection to the server
s = socket()
s.connect(("localhost", 8080))

# Compose the message/HTTP request we want to send to the server
msgPart1 = b"GET /../.gitignore HTTP/1.1\r\nHost: Ha\r\n\r\n"

# Send out the request
s.sendall(msgPart1)

# Listen for response and print it out
print (s.recv(4096))
s.close()
