// ui.go
package main

import (
	"image/color"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/glfwcanvas"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/pdu"
)

var BitmapCH chan *Bitmap

func ui_paint_bitmap(b *Bitmap) {
	BitmapCH <- b
}
func initUI(g *Client, width, height int) {
	//runtime.LockOSThread()
	wnd, cv, err := glfwcanvas.CreateWindow(width, height, "Hello")
	if err != nil {
		panic(err)
	}
	defer wnd.Close()

	wnd.KeyUp = g.KeyUp
	wnd.KeyDown = g.KeyDown
	wnd.MouseMove = g.MouseMove
	wnd.MouseUp = g.MouseUp
	wnd.MouseDown = g.MouseDown
	wnd.SizeChange = g.SizeChange
	wnd.MainLoop(func() {
		//fmt.Println("FPS:", wnd.FPS())
		select {
		case b := <-BitmapCH:
			paint_bitmap(cv, b)
		default:
		}
	})

}

func paint_bitmap(cv *canvas.Canvas, b *Bitmap) {
	m := cv.GetImageData(0, 0, b.Width, b.Height)

	var c color.RGBA
	i := 0
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			c = color.RGBA{b.Data[i+2], b.Data[i+1], b.Data[i], 255}
			i += 4
			m.Set(x, y, c)
		}
	}
	cv.PutImageData(m, b.DestLeft, b.DestTop)
}
func (g *Client) KeyUp(sc int, r rune, name string) {
	glog.Info("KeyUp:", sc, "r:", r, "name:", name)

	p := &pdu.ScancodeKeyEvent{}
	p.KeyCode = uint16(r)
	p.KeyboardFlags |= pdu.KBDFLAGS_RELEASE
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_SCANCODE, []pdu.InputEventsInterface{p})
}
func (g *Client) KeyDown(sc int, r rune, name string) {
	glog.Info("KeyDown:", sc, "r:", r, "name:", name)

	p := &pdu.ScancodeKeyEvent{}
	p.KeyCode = uint16(r)
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_SCANCODE, []pdu.InputEventsInterface{p})
}

func (g *Client) MouseMove(x, y int) {
	glog.Info("MouseMove", x, ":", y)
	p := &pdu.PointerEvent{}
	p.PointerFlags |= pdu.PTRFLAGS_MOVE
	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_MOUSE, []pdu.InputEventsInterface{p})
}

func (g *Client) MouseUp(button int, x, y int) {
	glog.Info("MouseUp", x, ":", y, ":", button)
	p := &pdu.PointerEvent{}

	switch button {
	case 0:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON1
	case 1:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON2
	case 2:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON3
	default:
		p.PointerFlags |= pdu.PTRFLAGS_MOVE
	}

	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_MOUSE, []pdu.InputEventsInterface{p})
}
func (g *Client) MouseDown(button int, x, y int) {
	glog.Info("MouseDown:", x, ":", y, ":", button)
	p := &pdu.PointerEvent{}

	p.PointerFlags |= pdu.PTRFLAGS_DOWN

	switch button {
	case 0:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON1
	case 1:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON2
	case 2:
		p.PointerFlags |= pdu.PTRFLAGS_BUTTON3
	default:
		p.PointerFlags |= pdu.PTRFLAGS_MOVE
	}

	p.XPos = uint16(x)
	p.YPos = uint16(y)
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_MOUSE, []pdu.InputEventsInterface{p})
}
func (g *Client) SizeChange(w, h int) {
}
