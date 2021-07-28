// rfb.go
package main

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/rfb"
)

type VncClient struct {
	Host   string // ip:port
	Width  int
	Height int
	vnc    *rfb.RFB
}

func NewVncClient(host string, width, height int, logLevel glog.LEVEL) *VncClient {
	return &VncClient{
		Host:   host,
		Width:  width,
		Height: height,
	}
}
func uiVnc(info *Info) (error, *VncClient) {
	BitmapCH = make(chan []Bitmap, 500)
	g := NewVncClient(fmt.Sprintf("%s:%s", info.Ip, info.Port), info.Width, info.Height, glog.INFO)

	g.Login()

	return nil, g
}

func (g *VncClient) Login() error {
	conn, err := net.DialTimeout("tcp", g.Host, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[dial err] %v", err)
	}
	//defer conn.Close()
	g.vnc = rfb.NewRFB(rfb.NewRFBConn(conn))

	g.vnc.On("error", func(e error) {
		glog.Info("on error")
		glog.Error(e)
	}).On("close", func() {
		err = errors.New("close")
		glog.Info("on close")
	}).On("success", func() {
		err = nil
		glog.Info("on success")
	}).On("ready", func() {
		glog.Info("on ready")
	}).On("update", func(br *rfb.BitRect) {
		glog.Debug("on update:", br)
		bs := make([]Bitmap, 0, 50)
		for _, v := range br.Rects {
			b := Bitmap{int(v.Rect.X), int(v.Rect.Y), int(v.Rect.X + v.Rect.Width), int(v.Rect.Y + v.Rect.Height),
				int(v.Rect.Width), int(v.Rect.Height),
				Bpp(uint16(br.Pf.BitsPerPixel)), false, v.Data}
			bs = append(bs, b)
		}

		ui_paint_bitmap(bs)
	})
	return nil
}
func (g *VncClient) SetRequestedProtocol(p uint32) {

}
func (g *VncClient) KeyUp(sc int, name string) {
	glog.Debug("KeyUp:", sc, "name:", name)
	k := &rfb.KeyEvent{}
	k.Key = uint32(sc)
	g.vnc.SendKeyEvent(k)
}
func (g *VncClient) KeyDown(sc int, name string) {
	glog.Debug("KeyDown:", sc, "name:", name)
	k := &rfb.KeyEvent{}
	k.DownFlag = 1
	k.Key = uint32(sc)
	g.vnc.SendKeyEvent(k)
}

func (g *VncClient) MouseMove(x, y int) {
	if g == nil {
		return
	}
	glog.Info("MouseMove", x, ":", y)
	p := &rfb.PointerEvent{}
	time.Sleep(8 * time.Millisecond)
	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.vnc.SendPointEvent(p)
}

func (g *VncClient) MouseWheel(scroll, x, y int) {
	glog.Info("MouseWheel", x, ":", y)
}

func (g *VncClient) MouseUp(button int, x, y int) {
	glog.Info("MouseUp", x, ":", y, ":", button)
	p := &rfb.PointerEvent{}

	switch button {
	case 0:
		p.Mask = 1
	case 2:
		p.Mask = 1<<3 - 1
	case 1:
		p.Mask = 1<<2 - 1
	default:
		p.Mask = 0
	}
	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.vnc.SendPointEvent(p)
}
func (g *VncClient) MouseDown(button int, x, y int) {
	glog.Info("MouseDown:", x, ":", y, ":", button)
	p := &rfb.PointerEvent{}

	switch button {
	case 0:
		p.Mask = 1
	case 2:
		p.Mask = 1<<3 - 1
	case 1:
		p.Mask = 1<<2 - 1
	default:
		p.Mask = 0
	}

	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.MouseMove(x, y)
	g.vnc.SendPointEvent(p)
}

func (g *VncClient) Close() {
	if g.vnc != nil {
		g.vnc.Close()
	}
}
