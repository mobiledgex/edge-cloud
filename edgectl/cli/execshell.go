package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	webrtc "github.com/pion/webrtc/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func RunWebrtcShell(exchangeFunc func(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error)) error {
	// hard code config for now
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:stun.mobiledgex.net:19302",
					"turn:stun.mobiledgex.net:19302",
				},
			},
		},
	}

	// create a new peer connection
	peerConn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to establish peer connection, %v", err)
	}

	dataChan, err := peerConn.CreateDataChannel("data", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel, %v", err)
	}

	// Set stdin and Stdout to raw
	sinOldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() {
		terminal.Restore(int(os.Stdin.Fd()), sinOldState)
	}()
	soutOldState, err := terminal.MakeRaw(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	defer func() {
		terminal.Restore(int(os.Stdout.Fd()), soutOldState)
	}()

	done := make(chan bool, 1)
	wr := webrtcutil.NewDataChanWriter(dataChan)

	dataChan.OnOpen(func() {
		go func() {
			// send stdin to data channel
			io.Copy(wr, os.Stdin)
			// close data channel if input stream is closed
			dataChan.Close()
		}()
	})
	dataChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		os.Stdout.Write(msg.Data)
	})
	dataChan.OnClose(func() {
		done <- true
	})

	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return err
	}
	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	answer, err := exchangeFunc(offer)
	if err != nil {
		return err
	}

	err = peerConn.SetRemoteDescription(*answer)
	if err != nil {
		return err
	}

	// wait for connection to complete
	<-done
	return nil
}
