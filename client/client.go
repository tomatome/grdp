package client

import (
	"log"
	"os"

	"github.com/tomatome/grdp/glog"
)

func init() {
	logger := log.New(os.Stdout, "", 0)
	glog.SetLogger(logger)
}

type RemoteControl interface {
	Login(host, user, passwd string, width, height int) error
	Connect() error
	KeyUp(sc int, name string)
	KeyDown(sc int, name string)
	MouseMove(x, y int)
	MouseWheel(scroll, x, y int)
	MouseUp(button int, x, y int)
	MouseDown(button int, x, y int)
	On(event string, msg interface{})
	OnBitmap(handler func([]Bitmap))
	Close()
}

type Bitmap struct {
	DestLeft     int    `json:"destLeft"`
	DestTop      int    `json:"destTop"`
	DestRight    int    `json:"destRight"`
	DestBottom   int    `json:"destBottom"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	BitsPerPixel int    `json:"bitsPerPixel"`
	IsCompress   bool   `json:"isCompress"`
	Data         []byte `json:"data"`
}

func Bpp(bp uint16) (pixel int) {
	switch bp {
	case 15:
		pixel = 1

	case 16:
		pixel = 2

	case 24:
		pixel = 3

	case 32:
		pixel = 4

	default:
		glog.Error("invalid bitmap data format")
	}
	return
}
