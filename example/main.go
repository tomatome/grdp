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
	"sync"
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

var im = 0

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	socketIO()
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
	server, _ := socketio.NewServer(nil)
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
			//s := so
			glog.Info(time.Now(), "on update Bitmap:", len(rectangles))
			bs := make([]Bitmap, 0, len(rectangles))
			mu := sync.Mutex{}
			wg := &sync.WaitGroup{}
			for _, v := range rectangles {
				wg.Add(1)
				//go func(wg *sync.WaitGroup) {
				IsCompress := v.IsCompress()
				data := v.BitmapDataStream
				glog.Info("data:", data)
				b1 := Bitmap{v.DestLeft, v.DestTop, v.DestRight, v.DestBottom,
					v.Width, v.Height, v.BitsPerPixel, IsCompress, data}

				so.Emit("rdp-bitmap", []Bitmap{b1})
				if IsCompress {
					data = decompress(&v)
					IsCompress = false
				}

				glog.Info(IsCompress, v.BitsPerPixel)
				b := Bitmap{v.DestLeft, v.DestTop, v.DestRight, v.DestBottom,
					v.Width, v.Height, v.BitsPerPixel, IsCompress, data}

				mu.Lock()
				so.Emit("rdp-bitmap", []Bitmap{b})
				bs = append(bs, b)
				mu.Unlock()
				wg.Done()
				//os.Exit(0)
				//}(wg)
			}
			wg.Wait()
			//so.Emit("rdp-bitmap", bs)
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
		return errors.New(fmt.Sprintf("[dial err] %v", err))
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
		return errors.New(fmt.Sprintf("[x224 connect err] %v", err))
	}
	glog.Info("wait connect ok")
	return nil
}

type Bitmap struct {
	DestLeft     uint16 `json:"destLeft"`
	DestTop      uint16 `json:"destTop"`
	DestRight    uint16 `json:"destRight"`
	DestBottom   uint16 `json:"destBottom"`
	Width        uint16 `json:"width"`
	Height       uint16 `json:"height"`
	BitsPerPixel uint16 `json:"bitsPerPixel"`
	IsCompress   bool   `json:"isCompress"`
	Data         []byte `json:"data"`
}

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
	glog.Info(fName)
	input := bitmap.BitmapDataStream
	glog.Info(bitmap.Width, bitmap.Height)
	output := bitmap_decompress(input, bitmap.Width, bitmap.Height)

	glog.Info("input:", input)
	glog.Info("output:", output)
	//os.Exit(0)

	return output
}

func process_plane(in []byte, width, height int, out *[]byte, idx int) int {
	var (
		color     byte
		x         byte
		last_line *[]byte
	)

	indexh := 0
	i := 0
	j := idx
	k := 0
	for indexh < height {
		j += width*height*4 - (indexh+1)*width*4
		color = 0
		this_line := (*out)[j:]
		indexw := 0
		if last_line == nil {
			glog.Info("first", indexw, width)
			for indexw < width {
				code := in[i]
				i++
				k++
				replen := code & 0xf
				collen := (code >> 4) & 0xf
				revcode := (replen << 4) | collen
				glog.Info("first1", indexw, revcode, collen, replen)
				if (revcode < 47) && (revcode >= 16) {
					replen = revcode
					collen = 0
				}
				glog.Info("first2", collen)
				for collen > 0 {
					color = in[i]
					i++
					k++
					(*out)[j] = color
					glog.Info("(*out)[j]", j, (*out)[j])
					j += 4
					indexw++
					collen--
				}
				glog.Info("first3", replen)
				for replen > 0 {
					(*out)[j] = color
					glog.Info("(*out)[j]", j, (*out)[j])
					j += 4
					indexw++
					replen--
				}
			}
			//glog.Info("out:", *out)
		} else {
			glog.Info("indexw", indexw, width, k, len(in))
			for indexw < width && i < len(in)-1 {
				code := in[i]
				i++
				k++
				replen := code & 0xf
				collen := (code >> 4) & 0xf
				revcode := (replen << 4) | collen
				glog.Info("code2", code, replen, collen, revcode)
				if (revcode < 47) && (revcode >= 16) {
					replen = revcode
					collen = 0
				}
				glog.Info("2=", collen)
				for collen > 0 && j < len(*last_line)-1 {
					x = in[i]
					i++
					k++
					glog.Info("x=", x, x&1, x>>1)
					if x&1 != 0 {
						x = x >> 1
						x = x + 1
						color = -x
					} else {
						x = x >> 1
						color = x
					}
					glog.Info("color:", color)
					x = (*last_line)[indexw*4] + color
					glog.Info("(*out)[indexw*4]=", indexw*4, (*last_line)[indexw*4])
					(*out)[j] = x
					j += 4
					indexw++
					collen--
				}
				glog.Info("3=", replen)
				for replen > 0 && j < len(*last_line)-1 {
					x = (*last_line)[indexw*4] + color
					(*out)[j] = x
					glog.Infof("*out3[%d]=%v", j, (*out)[j])
					j += 4
					indexw++
					replen--
				}
			}
		}
		indexh++
		last_line = &this_line
	}
	return k - 1
}

func bitmap_decompress(input []byte, width1, height1 uint16) []byte {
	width, height := int(width1), int(height1)
	output := make([]byte, width*height*4)
	glog.Info(width, height, cap(output), len(input))
	code := input[0]
	if code != 0x10 {
		return nil
	}
	org := input
	total_pro := 1
	input0 := org[total_pro:]
	bytes_pro := process_plane(input0, width, height, &output, 3)
	glog.Info("output1:", output)
	glog.Info("total_pro:", total_pro, bytes_pro)
	total_pro += bytes_pro
	input0 = org[total_pro:]
	bytes_pro = process_plane(input0, width, height, &output, 2)
	glog.Info("total_pro:", total_pro, bytes_pro)
	total_pro += bytes_pro
	input0 = org[total_pro:]
	bytes_pro = process_plane(input0, width, height, &output, 1)
	glog.Info("total_pro:", total_pro, bytes_pro)
	total_pro += bytes_pro
	input0 = org[total_pro:]
	bytes_pro = process_plane(input0, width, height, &output, 0)
	glog.Info("total_pro:", total_pro, bytes_pro)
	total_pro += bytes_pro
	return output
}
