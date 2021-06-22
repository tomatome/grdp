// ui.go
package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"time"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	"github.com/google/gxui/themes/light"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/pdu"
)

var (
	gc            *Client
	driverc       gxui.Driver
	width, height int
)

func StartUI(w, h int) {
	width, height = w, h
	gl.StartDriver(appMain)
}
func appMain(driver gxui.Driver) {
	theme := light.CreateTheme(driver)
	window := theme.CreateWindow(width, height, "MSTSC")
	window.SetScale(flags.DefaultScaleFactor)
	img = theme.CreateImage()
	//img.SetMargin(math.Spacing{0, 0, 0, 10})
	ScreenImage = image.NewRGBA(image.Rect(0, 0, width, height))
	//texture := driver.CreateTexture(ScreenImage, 1)
	//img.SetTexture(texture)
	img.SetVisible(false)
	img.OnMouseDown(func(e gxui.MouseEvent) {
		gc.MouseDown(int(e.Button), e.Point.X, e.Point.Y)
	})
	img.OnMouseUp(func(e gxui.MouseEvent) {
		gc.MouseUp(int(e.Button), e.Point.X, e.Point.Y)
	})
	img.OnMouseMove(func(e gxui.MouseEvent) {
		gc.MouseMove(e.Point.X, e.Point.Y)
	})
	img.OnMouseScroll(func(e gxui.MouseEvent) {
		gc.MouseWheel(e.ScrollY, e.Point.X, e.Point.Y)
	})
	img.OnKeyDown(func(e gxui.KeyboardEvent) {
		fmt.Println("OnKeyDown:", e)
		gc.KeyDown(int(e.Key), "")
	})
	img.OnKeyUp(func(e gxui.KeyboardEvent) {
		fmt.Println("OnKeyUp:", e)
		gc.KeyUp(int(e.Key), "")
	})
	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetHorizontalAlignment(gxui.AlignCenter)
	layout.SetMargin(math.Spacing{0, 50, 0, 0})

	label := theme.CreateLabel()
	label.SetText("Welcome Mstsc")
	label.SetColor(gxui.Red)
	ip := theme.CreateTextBox()
	user := theme.CreateTextBox()
	passwd := theme.CreateTextBox()
	ip.SetDesiredWidth(width / 4)
	user.SetDesiredWidth(width / 4)
	passwd.SetDesiredWidth(width / 4)
	ip.SetText("192.168.0.132:6400")
	user.SetText("administrator")
	passwd.SetText("Jhadmin123")
	blayout := theme.CreateLinearLayout()
	blayout.SetDirection(gxui.LeftToRight)
	blayout.SetVerticalAlignment(gxui.AlignMiddle)
	blayout.SetMargin(math.Spacing{10, 10, 10, 10})

	statusBar := theme.CreateLabel()
	statusBar.SetSize(math.Size{width, 10})
	statusBar.SetText("test")
	statusBar.SetColor(gxui.Red)
	statusBar.SetMargin(math.Spacing{0, height - 10, 0, 0})
	bok := theme.CreateButton()
	bok.SetText("OK")
	bok.OnClick(func(e gxui.MouseEvent) {
		err, info := NewInfo(ip.Text(), user.Text(), passwd.Text())
		info.Width, info.Height = width, height
		if err != nil {
			statusBar.SetText(err.Error())
			return
		}
		driverc = driver
		err, gc = uiclient(info)
		if err != nil {
			statusBar.SetText(err.Error())
			return
		}
		layout.SetVisible(false)
		img.SetVisible(true)
	})
	bcancel := theme.CreateButton()
	bcancel.SetText("Clear")
	bcancel.OnClick(func(e gxui.MouseEvent) {
		ip.SetText("")
		user.SetText("")
		passwd.SetText("")
	})
	layout.AddChild(label)
	layout.AddChild(ip)
	layout.AddChild(user)
	layout.AddChild(passwd)
	blayout.AddChild(bok)
	blayout.AddChild(bcancel)
	layout.AddChild(blayout)
	window.AddChild(layout)
	window.AddChild(img)
	//window.AddChild(statusBar)
	window.OnClose(driver.Terminate)
	update()
}

var (
	ScreenImage *image.RGBA
	img         gxui.Image
)

func update() {
	go func() {
		for {
			select {
			case bs := <-BitmapCH:
				paint_bitmap(bs)
			default:
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func paint_bitmap(bs []Bitmap) {
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
	}
	driverc.Call(func() {
		texture := driverc.CreateTexture(ScreenImage, 1)
		img.SetTexture(texture)
	})

}

var BitmapCH chan []Bitmap

func ui_paint_bitmap(bs []Bitmap) {
	BitmapCH <- bs
}
func uiclient(info *Info) (error, *Client) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	BitmapCH = make(chan []Bitmap, 500)
	g := NewClient(fmt.Sprintf("%s:%s", info.Ip, info.Port), info.Width, info.Height, glog.INFO)
	err := g.Login(info.Domain, info.Username, info.Passwd)
	if err != nil {
		fmt.Println("Login:", err)
		return err, nil
	}

	g.pdu.On("error", func(e error) {
		fmt.Println("on error:", e)
	}).On("close", func() {
		err = errors.New("close")
		fmt.Println("on close")
	}).On("success", func() {
		fmt.Println("on success")
	}).On("ready", func() {
		fmt.Println("on ready")
	}).On("update", func(rectangles []pdu.BitmapData) {
		glog.Info(time.Now(), "on update Bitmap:", len(rectangles))
		bs := make([]Bitmap, 0, 50)
		for _, v := range rectangles {
			IsCompress := v.IsCompress()
			data := v.BitmapDataStream
			//glog.Info("data:", data)
			if IsCompress {
				data = decompress(&v)
				IsCompress = false
			}

			//glog.Info(IsCompress, v.BitsPerPixel)
			b := Bitmap{int(v.DestLeft), int(v.DestTop), int(v.DestRight), int(v.DestBottom),
				int(v.Width), int(v.Height), int(v.BitsPerPixel), IsCompress, data}
			//glog.Infof("b:%+v, %d==%d", b.DestLeft, len(b.Data), b.Width*b.Height*4)
			bs = append(bs, b)
		}
		ui_paint_bitmap(bs)
	})

	return nil, g
}
