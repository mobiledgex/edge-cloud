package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	webrtc "github.com/pion/webrtc/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var execApiCmd edgeproto.ExecApiClient

func NewExecRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "RunCommand",
		Short: "Run a command or shell on an AppInst",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecRequest(&gencmd.ExecRequestIn)
		},
	}
	cmd.Flags().AddFlagSet(gencmd.ExecRequestFlagSet)
	return cmd
}

func runExecRequest(req *edgeproto.ExecRequest) error {
	if execApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}

	// hard code config for now
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
				//URLs: []string{"stun:stun.mobiledgex.net:19302"},
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
	offerBytes, err := json.Marshal(&offer)
	if err != nil {
		return err
	}
	req.Offer = string(offerBytes)

	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	ctx := context.Background()
	reply, err := execApiCmd.RunCommand(ctx, req)
	if err != nil {
		return err
	}
	if reply.Err != "" {
		return fmt.Errorf("%s", reply.Err)
	}
	if reply.Answer == "" {
		return fmt.Errorf("empty answer")
	}

	answer := webrtc.SessionDescription{}
	err = json.Unmarshal([]byte(reply.Answer), &answer)
	if err != nil {
		return fmt.Errorf("unable to unmarshal answer %s, %v",
			reply.Answer, err)
	}

	err = peerConn.SetRemoteDescription(answer)
	if err != nil {
		return err
	}

	// wait for connection to complete
	<-done
	return nil
}
