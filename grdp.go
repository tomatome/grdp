package grdp

import (
	"errors"
	"fmt"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/sec"
	"github.com/icodeface/grdp/protocol/t125"
	"github.com/icodeface/grdp/protocol/tpkt"
	"github.com/icodeface/grdp/protocol/x224"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type GrdpClient struct {
	Host string // ip:port
	tpkt *tpkt.TPKT
	x224 *x224.X224
	mcs  *t125.MCSClient
	sec  *sec.Client
}

func NewClient(host string, logLevel glog.LEVEL) *GrdpClient {
	glog.SetLevel(logLevel)
	logger := log.New(os.Stdout, "", 0)
	glog.SetLogger(logger)
	return &GrdpClient{
		Host: host,
	}
}

func (g *GrdpClient) Login(user, pwd string) error {
	conn, err := net.DialTimeout("tcp", g.Host, 3*time.Second)
	if err != nil {
		return errors.New(fmt.Sprintf("[dial err] %v", err))
	}
	defer conn.Close()

	g.tpkt = tpkt.New(core.NewSocketLayer(conn))
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMCSClient(g.x224)
	g.sec = sec.NewClient(g.mcs)
	g.sec.SetUser(user)
	g.sec.SetPwd(pwd)
	g.sec.SetDomain(strings.Split(g.Host, ":")[0])

	err = g.x224.Connect()
	if err != nil {
		return errors.New(fmt.Sprintf("[x224 connect err] %v", err))
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	g.sec.On("error", func(e error) {
		err = e
		glog.Error(e)
		wg.Done()
	}).On("close", func() {
		err = errors.New("close")
		glog.Info("close")
		wg.Done()
	}).On("connect", func() {
		err = nil
		glog.Info("connect")
		wg.Done()
	})

	wg.Wait()
	return err
}
