package grdp

import (
	"github.com/icodeface/grdp/protocol/t125"
	"github.com/icodeface/grdp/protocol/tpkt"
	"github.com/icodeface/grdp/protocol/x224"
	"net"
)

type GrdpClient struct {
	Host string // ip:port
	tpkt *tpkt.TPKT
	x224 *x224.X224
	mcs  *t125.Mcs
}

func NewClient(host string) *GrdpClient {
	return &GrdpClient{
		Host: host,
	}
}

func (g *GrdpClient) Login(user, pwd string) error {
	conn, err := net.Dial("tcp", g.Host)
	if err != nil {
		return err
	}
	g.tpkt = tpkt.New(conn)
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMcs(g.x224)

	g.x224.Connect()

	return nil
}
