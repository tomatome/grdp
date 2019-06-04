package pdu

import (
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/emission"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/t125/gcc"
)

type PDU struct {
	emission.Emitter
	transport core.Transport
}

func NewPDU(t core.Transport) *PDU {
	p := &PDU{
		*emission.NewEmitter(),
		t,
	}

	t.On("close", func() {
		p.Emit("close")
	}).On("error", func(err error) {
		p.Emit("error", err)
	})
	return p
}

type Client struct {
	*PDU
	clientCoreData *gcc.ClientCoreData
	userId         uint16
}

func NewClient(t core.Transport) *Client {
	c := &Client{
		PDU: NewPDU(t),
	}
	c.transport.Once("connect", c.connect)
	return c
}

func (c *Client) connect(data *gcc.ClientCoreData, userId uint16) {
	glog.Debug("pdu connect")
	c.clientCoreData = data
	c.userId = userId
	c.transport.Once("data", c.recvDemandActivePDU)
}

func (c *Client) recvDemandActivePDU(s []byte) {
	glog.Debug("pdu recvDemandActivePDU")
}

func (c *Client) recvServerSynchronizePDU() {

}

func (c *Client) recvServerControlCooperatePDU() {

}

func (c *Client) recvPDU() {

}
