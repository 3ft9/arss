import sys
import zmq
import datetime
import requests
import time

address = "tcp://localhost:6543"
if len(sys.argv) > 1:
	address =  sys.argv[1]

context = zmq.Context()
socket = context.socket(zmq.SUB)

print "Connected to ARSS..."
socket.connect(address)

socket.setsockopt(zmq.SUBSCRIBE, "")

next_upload = datetime.datetime.utcnow()
payload = ""
while 1:
	try:
		print socket.recv(zmq.NOBLOCK)
		#payload += socket.recv(zmq.NOBLOCK) + "\r\n"
		#print "Received item"
	except zmq.ZMQError, e:
		if e.errno == zmq.EAGAIN:
			pass  # no message was ready
		else:
			raise  # real error

	# if next_upload < datetime.datetime.utcnow():
	# 	if len(payload) > 0:
	# 		print "Uploading...",
	# 		url = "https://in.stagingdatasift.com/077dd1c73fde4f21b4aa4e4e713a5de4"
	# 		headers = {'Content-type': 'application/json', 'Auth': 'stuartdallas:78b8b0a3d933ba52a671b03dc35d783f'}
	# 		r = requests.post(url, data=payload, headers=headers, verify=False)
	# 		payload = ""
	# 		print "Done"
	# 	next_upload = datetime.datetime.utcnow() + datetime.timedelta(seconds=15)
	# else:
	# 	time.sleep(5)
