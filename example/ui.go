// ui.go
package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/glfwcanvas"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/pdu"
)

var BitmapCH chan []Bitmap

var ScreenImage *image.RGBA

func ui_paint_bitmap(bs []Bitmap) {
	BitmapCH <- bs
}

func initUI(g *Client, width, height int) {
	ScreenImage = image.NewRGBA(image.Rect(0, 0, width, height))

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
	wnd.MouseWheel = g.MouseWheel

	wnd.MainLoop(func() {
		select {
		case bs := <-BitmapCH:
			paint_bitmap(cv, bs)
		default:
		}
	})

}

func paint_bitmap(cv *canvas.Canvas, bs []Bitmap) {
	for _, b := range bs {
		m := image.NewRGBA(image.Rect(0, 0, b.Width, b.Height))
		i := 0
		for y := 0; y < b.Height; y++ {
			for x := 0; x < b.Width; x++ {
				c := color.RGBA{b.Data[i+2], b.Data[i+1], b.Data[i], 255}
				i += 4
				m.Set(x, y, c)
			}
		}
		draw.Draw(ScreenImage, ScreenImage.Bounds().Add(image.Pt(b.DestLeft, b.DestTop)), m, m.Bounds().Min, draw.Src)
		//cv.ClearRect(float64(b.DestLeft), float64(b.DestTop), float64(b.Width), float64(b.Height))
		//cv.PutImageData(m, b.DestLeft, b.DestTop)
	}

	cv.ClearRect(float64(0), float64(0), float64(cv.Width()), float64(cv.Height()))
	cv.PutImageData(ScreenImage, 0, 0)
}

func (g *Client) KeyUp(sc int, r rune, name string) {
	glog.Info("KeyUp:", sc, "r:", r, "name:", name)

	p := &pdu.ScancodeKeyEvent{}
	p.KeyCode = uint16(sc)
	p.KeyboardFlags |= pdu.KBDFLAGS_RELEASE
	g.pdu.SendInputEvents(pdu.INPUT_EVENT_SCANCODE, []pdu.InputEventsInterface{p})
}
func (g *Client) KeyDown(sc int, r rune, name string) {
	glog.Info("KeyDown:", sc, "r:", r, "name:", name)

	p := &pdu.ScancodeKeyEvent{}
	p.KeyCode = uint16(sc)
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

func (g *Client) MouseWheel(x, y int) {
	glog.Info("MouseWheel", x, ":", y)
	p := &pdu.PointerEvent{}
	p.PointerFlags |= pdu.PTRFLAGS_WHEEL
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
