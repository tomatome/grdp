// main.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"text/template"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/tomatome/grdp/core"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/nla"
	"github.com/tomatome/grdp/protocol/pdu"
	"github.com/tomatome/grdp/protocol/sec"
	"github.com/tomatome/grdp/protocol/t125"
	"github.com/tomatome/grdp/protocol/tpkt"
	"github.com/tomatome/grdp/protocol/x224"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//web example
	//socketIO()
	//client example
	uiclient()
}
func showPreview(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/html/index.html")
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	w.Header().Add("Content-Type", "text/html")
	t.Execute(w, nil)

}

type Screen struct {
	Height uint16 `json:"height"`
	Width  uint16 `json:"width"`
}

type Info struct {
	Domain   string `json:"domain"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Passwd   string `json:"password"`
	Screen   `json:"screen"`
}

func socketIO() {
	server := socketio.NewServer(nil)
	server.OnConnect("/", func(so socketio.Conn) error {
		fmt.Println("OnConnect", so.ID())
		so.Emit("rdp-connect", true)
		return nil
	})
	server.OnEvent("/", "infos", func(so socketio.Conn, data interface{}) {
		var info Info
		v, _ := json.Marshal(data)
		json.Unmarshal(v, &info)
		fmt.Println(so.ID(), "logon infos:", info)

		g := NewClient(fmt.Sprintf("%s:%s", info.Ip, info.Port), glog.INFO)
		err := g.Login(info.Domain, info.Username, info.Passwd, info.Width, info.Height)
		if err != nil {
			fmt.Println("Login:", err)
			so.Emit("rdp-error", "{\"code\":1,\"message\":\""+err.Error()+"\"}")
			return
		}
		so.SetContext(g)
		g.pdu.On("error", func(e error) {
			fmt.Println("on error:", e)
			so.Emit("rdp-error", "{\"code\":1,\"message\":\""+e.Error()+"\"}")
			//wg.Done()
		}).On("close", func() {
			err = errors.New("close")
			fmt.Println("on close")
		}).On("success", func() {
			fmt.Println("on success")
		}).On("ready", func() {
			fmt.Println("on ready")
		}).On("update", func(rectangles []pdu.BitmapData) {
			glog.Info(time.Now(), "on update Bitmap:", len(rectangles))
			bs := make([]Bitmap, 0, len(rectangles))
			for _, v := range rectangles {
				IsCompress := v.IsCompress()
				data := v.BitmapDataStream
				glog.Debug("data:", data)
				if IsCompress {
					data = decompress(&v)
					IsCompress = false
				}

				glog.Debug(IsCompress, v.BitsPerPixel)
				b := Bitmap{int(v.DestLeft), int(v.DestTop), int(v.DestRight), int(v.DestBottom),
					int(v.Width), int(v.Height), int(v.BitsPerPixel), IsCompress, data}
				so.Emit("rdp-bitmap", []Bitmap{b})
				bs = append(bs, b)
			}
			so.Emit("rdp-bitmap", bs)
		})
	})

	server.OnEvent("/", "mouse", func(so socketio.Conn, x, y uint16, button int, isPressed bool) {
		glog.Info("mouse", x, ":", y, ":", button, ":", isPressed)
		p := &pdu.PointerEvent{}
		if isPressed {
			p.PointerFlags |= pdu.PTRFLAGS_DOWN
		}

		switch button {
		case 1:
			p.PointerFlags |= pdu.PTRFLAGS_BUTTON1
		case 2:
			p.PointerFlags |= pdu.PTRFLAGS_BUTTON2
		case 3:
			p.PointerFlags |= pdu.PTRFLAGS_BUTTON3
		default:
			p.PointerFlags |= pdu.PTRFLAGS_MOVE
		}

		p.XPos = x
		p.YPos = y
		g := so.Context().(*Client)
		g.pdu.SendInputEvents(pdu.INPUT_EVENT_MOUSE, []pdu.InputEventsInterface{p})
	})

	//keyboard
	server.OnEvent("/", "scancode", func(so socketio.Conn, button uint16, isPressed bool) {
		glog.Info("scancode:", "button:", button, "isPressed:", isPressed)

		p := &pdu.ScancodeKeyEvent{}
		p.KeyCode = button
		if !isPressed {
			p.KeyboardFlags |= pdu.KBDFLAGS_RELEASE
		}
		g := so.Context().(*Client)
		g.pdu.SendInputEvents(pdu.INPUT_EVENT_SCANCODE, []pdu.InputEventsInterface{p})

	})

	//wheel
	server.OnEvent("/", "wheel", func(so socketio.Conn, x, y, step uint16, isNegative, isHorizontal bool) {
		glog.Info("wheel", x, ":", y, ":", step, ":", isNegative, ":", isHorizontal)
		var p = &pdu.PointerEvent{}
		if isHorizontal {
			p.PointerFlags |= pdu.PTRFLAGS_HWHEEL
		} else {
			p.PointerFlags |= pdu.PTRFLAGS_WHEEL
		}

		if isNegative {
			p.PointerFlags |= pdu.PTRFLAGS_WHEEL_NEGATIVE
		}

		p.PointerFlags |= (step & pdu.WheelRotationMask)
		p.XPos = x
		p.YPos = y
		g := so.Context().(*Client)
		g.pdu.SendInputEvents(pdu.INPUT_EVENT_SCANCODE, []pdu.InputEventsInterface{p})
	})

	server.OnError("/", func(so socketio.Conn, err error) {
		if so == nil || so.Context() == nil {
			return
		}
		fmt.Println("error:", err)
		g := so.Context().(*Client)
		if g != nil {
			g.tpkt.Close()
		}
		so.Close()
	})

	server.OnDisconnect("/", func(so socketio.Conn, s string) {
		if so.Context() == nil {
			return
		}
		fmt.Println("OnDisconnect:", s)
		so.Emit("rdp-error", "{code:1,message:"+s+"}")

		g := so.Context().(*Client)
		if g != nil {
			g.tpkt.Close()
		}
		so.Close()
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/css/", http.FileServer(http.Dir("static")))
	http.Handle("/js/", http.FileServer(http.Dir("static")))
	http.Handle("/img/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/", showPreview)

	log.Println("Serving at localhost:8088...")
	log.Fatal(http.ListenAndServe(":8088", nil))
}

type Client struct {
	Host string // ip:port
	tpkt *tpkt.TPKT
	x224 *x224.X224
	mcs  *t125.MCSClient
	sec  *sec.Client
	pdu  *pdu.Client
}

func NewClient(host string, logLevel glog.LEVEL) *Client {
	glog.SetLevel(logLevel)
	logger := log.New(os.Stdout, "", 0)
	glog.SetLogger(logger)
	return &Client{
		Host: host,
	}
}
func (g *Client) Login(domain, user, pwd string, width, height uint16) error {
	glog.Info("Connect:", g.Host, "with", domain+"\\"+user, ":", pwd)
	conn, err := net.DialTimeout("tcp", g.Host, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[dial err] %v", err)
	}
	//defer conn.Close()

	g.tpkt = tpkt.New(core.NewSocketLayer(conn), nla.NewNTLMv2(domain, user, pwd))
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMCSClient(g.x224)
	g.sec = sec.NewClient(g.mcs)
	g.pdu = pdu.NewClient(g.sec)

	g.mcs.SetClientCoreData(width, height)

	g.sec.SetUser(user)
	g.sec.SetPwd(pwd)
	g.sec.SetDomain(domain)

	g.tpkt.SetFastPathListener(g.sec)
	g.sec.SetFastPathListener(g.pdu)
	g.pdu.SetFastPathSender(g.tpkt)

	//g.x224.SetRequestedProtocol(x224.PROTOCOL_RDP)
	g.x224.SetRequestedProtocol(x224.PROTOCOL_SSL)

	err = g.x224.Connect()
	if err != nil {
		return fmt.Errorf("[x224 connect err] %v", err)
	}
	glog.Info("wait connect ok")
	return nil
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
type ByteSlice []byte

func (x ByteSlice) Len() int           { return len(x) }
func (x ByteSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x ByteSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func decompress(bitmap *pdu.BitmapData) []byte {
	var fName string
	switch bitmap.BitsPerPixel {
	case 15:
		fName = "bitmap_decompress_15"

	case 16:
		fName = "bitmap_decompress_16"

	case 24:
		fName = "bitmap_decompress_24"

	case 32:
		fName = "bitmap_decompress_32"

	default:
		glog.Error("invalid bitmap data format")
	}
	glog.Debug(fName)
	input := bitmap.BitmapDataStream
	glog.Error(bitmap.Width, bitmap.Height)
	glog.Error("input:")
	for _, v := range input {
		fmt.Printf("%d,", v)
	}
	fmt.Printf("\n")
	output := bitmap_decompress(input, int(bitmap.Width), int(bitmap.Height), 4)

	//sort.Reverse(ByteSlice(output))
	glog.Error("output:", output)
	return output
}
func init() {
	BitmapCH = make(chan []Bitmap, 100)

}
func uiclient() {
	screen := Screen{800, 1000}
	info := Info{".", "192.168.0.132", "6400", "administrator", "Jhadmin123", screen}
	g := NewClient(fmt.Sprintf("%s:%s", info.Ip, info.Port), glog.ERROR)
	err := g.Login(info.Domain, info.Username, info.Passwd, info.Width, info.Height)
	if err != nil {
		fmt.Println("Login:", err)
		return
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
		bs := make([]Bitmap, 0, len(rectangles))
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
	initUI(g, int(screen.Width), int(screen.Height))
	time.Sleep(10000 * time.Second)
}
