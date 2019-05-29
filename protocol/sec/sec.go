package sec

import (
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/core"
)

type SEC struct {
	emission.Emitter
	transport core.Transport
}

func NewSEC(t core.Transport) *SEC {
	sec := &SEC{
		*emission.NewEmitter(),
		t,
	}

	t.On("close", func() {
		sec.Emit("close")
	}).On("error", func(err error) {
		sec.Emit("error", err)
	})

	return sec
}

type Client struct {
	*SEC
}

func NewClient(t core.Transport) *Client {
	return &Client{
		NewSEC(t),
	}
}

func (c *Client) sendInfoPkt() {

}
