package grdp

import (
	"errors"
	"fmt"
	"github.com/icodeface/grdp/protocol/t125"
	"github.com/icodeface/grdp/protocol/tpkt"
	"github.com/icodeface/grdp/protocol/x224"
	"net"
	"time"
)

type GrdpClient struct {
	Host string // ip:port
	tpkt *tpkt.TPKT
	x224 *x224.X224
	mcs  *t125.MCS
}

func NewClient(host string) *GrdpClient {
	return &GrdpClient{
		Host: host,
	}
}

func (g *GrdpClient) Login(user, pwd string) error {
	conn, err := net.Dial("tcp", g.Host)
	if err != nil {
		return errors.New(fmt.Sprintf("[dial err] %v", err))
	}

	g.tpkt = tpkt.New(conn)
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMCS(g.x224, t125.SEND_DATA_INDICATION, t125.SEND_DATA_REQUEST)

	err = g.x224.Connect()
	if err != nil {
		return errors.New(fmt.Sprintf("[x224 connect err] %v", err))
	}

	time.Sleep(15 * time.Second)
	return nil
}
