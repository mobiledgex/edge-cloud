package proxyutil

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"

	"github.com/xtaci/smux"
)

func ProxyMuxServer(serverConn net.Conn, remoteUrl string) error {
	urlObj, err := url.Parse(remoteUrl)
	if err != nil {
		return fmt.Errorf("failed to parse console url %s, %v", remoteUrl, err)
	}
	sess, err := smux.Server(serverConn, nil)
	if err != nil {
		return fmt.Errorf("failed to setup smux server, %v", err)
	}
	defer sess.Close()
	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			if err.Error() != io.ErrClosedPipe.Error() {
				return fmt.Errorf("failed to setup smux acceptstream, %v", err)
			}
			return nil
		}
		var server net.Conn
		if urlObj.Scheme == "http" {
			server, err = net.Dial("tcp", urlObj.Host)
			if err != nil {
				return fmt.Errorf("failed to get console, %v", err)
			}
		} else if urlObj.Scheme == "https" {
			server, err = tls.Dial("tcp", urlObj.Host, &tls.Config{
				InsecureSkipVerify: true,
			})
			if err != nil {
				return fmt.Errorf("failed to get console, %v", err)
			}
		} else {
			return fmt.Errorf("unsupported scheme %s", urlObj.Scheme)
		}
		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					break
				}
				server.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)

		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := server.Read(buf)
				if err != nil {
					break
				}
				stream.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)
	}
}

func ProxyMuxClient(entryListener net.Listener, clientConn net.Conn) error {
	sess, err := smux.Client(clientConn, nil)
	if err != nil {
		return fmt.Errorf("failed to setup smux server: %v", err)
	}
	for {
		entryConn, err := entryListener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connections from %s, %v", entryListener.Addr().String(), err)
		}

		stream, err := sess.OpenStream()
		if err != nil {
			return fmt.Errorf("failed to open smux stream, %v", err)
		}
		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					break
				}
				server.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(entryConn, stream)

		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := server.Read(buf)
				if err != nil {
					break
				}
				stream.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(entryConn, stream)
	}
}
