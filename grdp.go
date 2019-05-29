package grdp

import (
	"errors"
	"fmt"
	"github.com/icodeface/grdp/core"
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
	mcs  *t125.MCSClient
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
	defer conn.Close()

	g.tpkt = tpkt.New(core.NewSocketLayer(conn))
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMCSClient(g.x224)

	err = g.x224.Connect()
	if err != nil {
		return errors.New(fmt.Sprintf("[x224 connect err] %v", err))
	}

	g.mcs.On("error", func(err error) {
		fmt.Println(err)
	})

	time.Sleep(15 * time.Second)
	return nil
}
