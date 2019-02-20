IP test traffic tool.  Currently tests UDP which is the most problematic.

Runs in either server or client mode.  In server mode (--mode server) it will listen on UDP port 4444 and send a response
for every message received.  The response will be the same contents and length as the incoming request.

In client mode (--mode client) it will send a quantity of messages equivalent to --loop of a size --size.   The client
will count the number of responses received round trip.

Server mode is the default and can be run with no arguments, or in Kubernetes.  Port 4444 must be open.

Client mode example
docker run --network=host registry.mobiledgex.net:5000/mobiledgex/iptest   --mode client --address 37.50.143.119:4444  --size 2000 --loops 100
client packet-written: bytes=2000
2019/02/20 14:55:27 client packet-received: bytes=2000 from=37.50.143.119:4444 sent=1 rcvd=1
client packet-written: bytes=2000
2019/02/20 14:55:27 client packet-received: bytes=2000 from=37.50.143.119:4444 sent=2 rcvd=2
client packet-written: bytes=2000
...
...
2019/02/20 20:58:02 Done: size=2000 sent=100 received=100 fail=0

Usage of ./iptest:
  -address string
        dest address of server for client
  -loops int
        client loops
  -mode string
        mode server or client (default "server")
  -size int
        client packet size (default -1)
