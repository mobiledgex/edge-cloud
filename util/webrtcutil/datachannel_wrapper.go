package webrtcutil

import (
	"io"
	"net"
	"os"
	"time"

	webrtc "github.com/pion/webrtc/v2"
)

type dataChannelAddr struct{}

func (addr dataChannelAddr) Network() string {
	return "mobiledgex-webrtc"
}

func (addr dataChannelAddr) String() string {
	return "webrtc://mobiledgex-datachannel"
}

// A DataChannel implements the net.Conn interface over a webrtc data channel
type DataChannel struct {
	dc *webrtc.DataChannel
	rr io.ReadCloser
}

var _ net.Conn = (*DataChannel)(nil)

// WrapDataChannel wraps an rtc data channel and implements the net.Conn
// interface
func WrapDataChannel(rtcDataChannel *webrtc.DataChannel) (*DataChannel, error) {
	rr, rw, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	dc := &DataChannel{
		dc: rtcDataChannel,
		rr: rr,
	}
	dc.dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		if rw != nil {
			_, err := rw.Write(msg.Data)
			if err != nil {
				rw.Close()
				rw = nil
			}
		}
	})
	return dc, nil
}

func (dc *DataChannel) Read(b []byte) (n int, err error) {
	return dc.rr.Read(b)
}

func (dc *DataChannel) Write(b []byte) (n int, err error) {
	err = dc.dc.Send(b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (dc *DataChannel) Close() error {
	dc.rr.Close()
	return nil
}

func (dc *DataChannel) LocalAddr() net.Addr {
	return dataChannelAddr{}
}

func (dc *DataChannel) RemoteAddr() net.Addr {
	return dataChannelAddr{}
}

func (dc *DataChannel) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (dc *DataChannel) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (dc *DataChannel) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}
