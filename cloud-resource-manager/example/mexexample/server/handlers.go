package server

import (
	"fmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample/api"
	"golang.org/x/net/context"
	"net"
	"os"
)

type Server struct{}

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
		Message:    req.Message,
		Outbound:   GetOutboundIP(),
		Interfaces: []*api.Interface{},
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		res.Message += fmt.Sprintf("error getting interfaces, %v ", err)
	}
	hn, err := os.Hostname()
	if err != nil {
		res.Message += fmt.Sprintf("error getting hostname, %v ", err)
	} else {
		res.Hostname = hn
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
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%v", localAddr.IP)
}
