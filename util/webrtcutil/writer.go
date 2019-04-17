package webrtcutil

import webrtc "github.com/pion/webrtc/v2"

type DataChanWriter struct {
	dataChan *webrtc.DataChannel
}

func NewDataChanWriter(d *webrtc.DataChannel) *DataChanWriter {
	return &DataChanWriter{dataChan: d}
}

func (s *DataChanWriter) Write(p []byte) (int, error) {
	err := s.dataChan.Send(p)
	return len(p), err
}
