IP test traffic tool.  Currently tests UDP which is the most problematic.

Runs in either server or client mode.  In server mode (--mode server) it will listen on UDP port 4444 and send a response
for every message received.  The response will be the same contents and length as the incoming request.

In client mode (--mode client) it will send a quantity of messages equivalent to --loop of a size --size.   The client
will count the number of responses received round trip.

Server mode is the default and can be run with no arguments, or in Kubernetes.  Port 4444 must be open.

Client mode example
RRIS-MAC:~ jimmorris$ docker run --network=host registry.mobiledgex.net:5000/mobiledgex/iptest --mode client --address 80.187.128.28:4444 --size 200 --loops 10
2019/02/20 22:43:06 client mode
2019/02/20 22:43:06 client sent -- bytes: 200
2019/02/20 22:43:07 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 1 rcvd: 1
2019/02/20 22:43:07 client sent -- bytes: 200
2019/02/20 22:43:07 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 2 rcvd: 2
2019/02/20 22:43:07 client sent -- bytes: 200
2019/02/20 22:43:07 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 3 rcvd: 3
2019/02/20 22:43:07 client sent -- bytes: 200
2019/02/20 22:43:07 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 4 rcvd: 4
2019/02/20 22:43:07 client sent -- bytes: 200
2019/02/20 22:43:07 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 5 rcvd: 5
2019/02/20 22:43:07 client sent -- bytes: 200
2019/02/20 22:43:08 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 6 rcvd: 6
2019/02/20 22:43:08 client sent -- bytes: 200
2019/02/20 22:43:08 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 7 rcvd: 7
2019/02/20 22:43:08 client sent -- bytes: 200
2019/02/20 22:43:08 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 8 rcvd: 8
2019/02/20 22:43:08 client sent -- bytes: 200
2019/02/20 22:43:08 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 9 rcvd: 9
2019/02/20 22:43:08 client sent -- bytes: 200
2019/02/20 22:43:08 client received -- bytes: 200 from: 80.187.128.28:4444 sent: 10 rcvd: 10
2019/02/20 22:43:08 
Done -- size: 200 sent: 10 received: 10 fail: 0

Usage of ./iptest:
  -address string
        dest address of server for client
  -loops int
        client loops
  -mode string
        mode server or client (default "server")
  -size int
        client packet size (default -1)
