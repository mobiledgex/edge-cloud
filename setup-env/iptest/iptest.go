package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	mode                = flag.String("mode", "server", "mode server or client")
	loops               = flag.Int("loops", 0, "client loops")
	size                = flag.Int("size", -1, "client packet size")
	address             = flag.String("address", "", "dest address of server for client")
	timeout             = time.Second * 5
	rate                = flag.Int("rate", 100, "rate in mseconds between requests")
	clientPktsReceived  = 0
	clientReceiveFailed = false
)

func clientSendPackets(conn *net.UDPConn, count int, size int, rate time.Duration) {

	// send a message of the requested size at the rate
	b := []byte("Z")
	data := string(bytes.Repeat(b, size))

	numSent := 0
	throttle := time.Tick(rate)
	for {
		<-throttle
		go fmt.Fprintf(conn, data)
		numSent++
		if numSent > (clientPktsReceived + 10) {
			// we are too far ahead of the receive, let's skip this interval
			log.Printf("** Waiting ** numSent: %d clientPktsReceived: %d\n", numSent, clientPktsReceived)
		} else {
			log.Printf("client sent -- bytes: %d sent: %d rcvd: %d\n", size, numSent, clientPktsReceived)
			if numSent >= count {
				break
			}
			if clientReceiveFailed {
				log.Printf("sending stopped due to receive failure")
				break
			}
		}
	}
}

func runClient(ctx context.Context, loops int, size int) error {
	raddr, err := net.ResolveUDPAddr("udp", *address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	go clientSendPackets(conn, loops, size, time.Millisecond*time.Duration(*rate))

	defer conn.Close()
	ch := make(chan error, 1)

	go func() {
		buffer := make([]byte, 100*1024)
		if err != nil {
			log.Printf("Error on open file %v", err)
			return
		}
		for i := 0; i < loops; i++ {
			deadline := time.Now().Add(timeout)
			err = conn.SetReadDeadline(deadline)
			if err != nil {
				ch <- err
				return
			}
			nRead, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				log.Printf("Failure on receive loop: %d -- %v\n", i, err)
				clientReceiveFailed = true
				ch <- err
			} else {
				clientPktsReceived++
				log.Printf("client received -- bytes: %d from: %s sent: %d rcvd: %d\n",
					nRead, addr.String(), loops, clientPktsReceived)
			}
		}
		ch <- nil

	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}
	log.Printf("Done -- size: %d sent: %d received: %d", size, loops, clientPktsReceived)
	return err
}

func runServer(ctx context.Context) error {
	const maxBufferSize = 100 * 1024 //100k buffer
	conn, err := net.ListenPacket("udp", "0.0.0.0:4444")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)
	totalPkts := 0
	go func() {
		for {
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				ch <- err
				return
			}

			totalPkts++
			log.Printf("server received -- bytes: %d from: %s total: %d\n",
				n, addr.String(), totalPkts)

			deadline := time.Now().Add(timeout)
			err = conn.SetWriteDeadline(deadline)
			if err != nil {
				ch <- err
				return
			}

			// echo the data back
			n, err = conn.WriteTo(buffer[:n], addr)
			if err != nil {
				ch <- err
				return
			}
			log.Printf("server sent -- bytes: %d to: %s\n", n, addr.String())
		}
	}()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}
	return err
}

func run() {
	ctx := context.Background()
	if *mode == "server" {
		log.Println("server mode")
		err := runServer(ctx)
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	} else if *mode == "client" {
		log.Println("client mode")
		err := runClient(ctx, *loops, *size)
		if err != nil {
			log.Fatalf("client error: %v", err)
		}
	} else {
		log.Fatalf("invalid mode %s", *mode)
	}

}

func validateArgs() {
	flag.Parse()
	if *mode == "client" {
		if *address == "" {
			log.Fatalf("need --address of server")
		}
		if *size == -1 {
			log.Fatalf("need --size of packet")
		}
		if *loops == 0 {
			log.Fatalf("need --loops for test")
		}
	}
}

func main() {
	validateArgs()
	run()
}
