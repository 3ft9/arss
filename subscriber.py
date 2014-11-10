import sys
import zmq

address = "tcp://127.0.0.1:6543"
if len(sys.argv) > 1:
	address =  sys.argv[1]

context = zmq.Context()
socket = context.socket(zmq.SUB)

print "Connected to ARSS..."
socket.connect(address)

socket.setsockopt(zmq.SUBSCRIBE, "")

while 1:
	print socket.recv()
