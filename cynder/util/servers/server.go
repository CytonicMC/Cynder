package servers

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"net"
)

type ServerInfoImpl struct {
	proxy.ServerInfo
	name string
	addr net.Addr
}

func (c *ServerInfoImpl) Name() string {
	return c.name
}

func (c *ServerInfoImpl) Addr() net.Addr {
	return c.addr
}

func CreateServerInfo(ip string, port int, id string) *ServerInfoImpl {
	return &ServerInfoImpl{
		name: id,
		addr: &net.TCPAddr{IP: net.ParseIP(ip), Port: port},
	}
}
