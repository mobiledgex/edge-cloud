package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample/api"
	"golang.org/x/net/context"
)

type Server struct{}

func (srv *Server) DataTest(ctx context.Context, req *api.DataTestRequest) (res *api.DataTestResponse, err error) {
	b := []byte("Z")
	response := string(bytes.Repeat(b, int(req.Numbytes)))
	res = &api.DataTestResponse{
		Bytes: response,
	}
	return res, nil
}

func (srv *Server) Echo(ctx context.Context, req *api.EchoRequest) (res *api.EchoResponse, err error) {
	res = &api.EchoResponse{
		Message: req.Message,
	}
	return res, nil
}

func (srv *Server) Status(ctx context.Context, req *api.StatusRequest) (res *api.StatusResponse, err error) {
	res = &api.StatusResponse{
		Message: req.Message,
		Status:  "OK",
	}
	return res, nil
}

func (srv *Server) Info(ctx context.Context, req *api.InfoRequest) (res *api.InfoResponse, err error) {
	res = &api.InfoResponse{
		Totaludp:     TotalUDP,
		Totaltcp:     TotalTCP,
		Message:      req.Message,
		Outbound:     GetOutboundIP(),
		Realoutbound: GetRealOutboundIP(),
		Interfaces:   []*api.Interface{},
	}
	hn, err := os.Hostname()
	if err != nil {
		res.Message += fmt.Sprintf("error getting hostname, %v ", err)
	} else {
		res.Hostname = hn
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		res.Message += fmt.Sprintf("error getting interfaces, %v ", err)
	}
	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			res.Message += fmt.Sprintf("error getting addrs for intf %v, %v ", intf, err)
		} else {
			intfAddrs := fmt.Sprintf("%v", addrs)
			res.Interfaces = append(res.Interfaces, &api.Interface{Name: intf.Name, Addresses: intfAddrs})
		}
	}
	return res, nil
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Sprintf("error getting outbound ip, %v", err)
	}
	defer conn.Close() // no lint
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%v", localAddr.IP)
}

func GetRealOutboundIP() string {
	resp, err := http.Get("http://api.ipify.org")
	if err != nil {
		fmt.Println("error getting real outbound ip")
		return "unavailable"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading resp body", err)
		return "unavailable"
	}
	return fmt.Sprintf("%s", body)
}
