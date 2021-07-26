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
		glog.Info("on update:", br)
		bs := make([]Bitmap, 0, 50)
		b := Bitmap{int(br.Rect.X), int(br.Rect.Y), int(br.Rect.X + br.Rect.Width), int(br.Rect.Y + br.Rect.Height),
			4, 3, //int(br.Rect.Width), int(br.Rect.Height),
			Bpp(uint16(br.Pf.BitsPerPixel)), false, br.Data}
		glog.Infof("b:%+v, %d==%d", b.DestLeft, len(b.Data), b.Width*b.Height*4)
		bs = append(bs, b)
		ui_paint_bitmap(bs)
	})
	return nil
}
func (g *VncClient) SetRequestedProtocol(p uint32) {

}
func (g *VncClient) KeyUp(sc int, name string) {
	glog.Debug("KeyUp:", sc, "name:", name)

}
func (g *VncClient) KeyDown(sc int, name string) {
	glog.Debug("KeyDown:", sc, "name:", name)

}

func (g *VncClient) MouseMove(x, y int) {
	glog.Debug("MouseMove", x, ":", y)

}

func (g *VncClient) MouseWheel(scroll, x, y int) {
	glog.Info("MouseWheel", x, ":", y)

}

func (g *VncClient) MouseUp(button int, x, y int) {
	glog.Debug("MouseUp", x, ":", y, ":", button)

}
func (g *VncClient) MouseDown(button int, x, y int) {
	glog.Info("MouseDown:", x, ":", y, ":", button)

}
