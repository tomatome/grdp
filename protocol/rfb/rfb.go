// rfb.go
package rfb

import (
	"bytes"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/lunixbochs/struc"

	"github.com/tomatome/grdp/core"
	"github.com/tomatome/grdp/emission"
	"github.com/tomatome/grdp/glog"
)

// ProtocolVersion
const (
	RFB003003 = "RFB 003.003\n"
	RFB003007 = "RFB 003.007\n"
	RFB003008 = "RFB 003.008\n"
)

// SecurityType
const (
	SEC_INVALID uint8 = 0
	SEC_NONE    uint8 = 1
	SEC_VNC     uint8 = 2
)

type RFBConn struct {
	emission.Emitter
	// The Socket connection to the client
	Conn    net.Conn
	s       *ServerInit
	NbRect  uint16
	BitRect *BitRect
}
type BitRect struct {
	Rect *Rectangle
	Pf   *PixelFormat
	Data []byte
}

func NewRFBConn(s net.Conn) *RFBConn {
	fc := &RFBConn{
		Emitter: *emission.NewEmitter(),
		Conn:    s,
		BitRect: new(BitRect),
	}
	core.StartReadBytes(12, fc, fc.recvProtocolVersion)

	return fc
}
func (fc *RFBConn) Read(b []byte) (n int, err error) {
	return fc.Conn.Read(b)
}

func (fc *RFBConn) Write(data []byte) (n int, err error) {
	buff := &bytes.Buffer{}
	buff.Write(data)
	return fc.Conn.Write(buff.Bytes())
}
func (fc *RFBConn) Close() error {
	return fc.Conn.Close()
}
func (fc *RFBConn) recvProtocolVersion(s []byte, err error) {
	version := string(s)
	glog.Debug("RFBConn recvProtocolVersion", version, err)
	if err != nil {
		fc.Emit("error", err)
		return
	}
	fc.Emit("data", version)

	if version == RFB003003 {
		fc.Emit("error", fmt.Errorf("%s", "Not Support RFB003003"))
		return
		//core.StartReadBytes(4, fc, fc.recvSecurityServer)
	} else {
		core.StartReadBytes(1, fc, fc.checkSecurityList)
	}
}
func (fc *RFBConn) checkSecurityList(s []byte, err error) {
	r := bytes.NewReader(s)
	result, _ := core.ReadUInt8(r)
	glog.Debug("RFBConn recvSecurityList", result, err)

	core.StartReadBytes(int(result), fc, fc.recvSecurityList)
}
func (fc *RFBConn) recvSecurityList(s []byte, err error) {
	r := bytes.NewReader(s)
	secLevel := SEC_VNC
	for r.Len() > 0 {
		result, _ := core.ReadUInt8(r)
		if result == SEC_NONE || result == SEC_VNC {
			secLevel = result
			break
		}
	}

	glog.Debug("RFBConn recvSecurityList", secLevel, err)
	buff := &bytes.Buffer{}
	core.WriteUInt8(secLevel, buff)
	fc.Write(buff.Bytes())
	if secLevel == SEC_VNC {
		core.StartReadBytes(16, fc, fc.recvVNCChallenge)
	} else {
		core.StartReadBytes(4, fc, fc.recvSecurityResult)
	}

}

func fixDesKeyByte(val byte) byte {
	var newval byte = 0
	for i := 0; i < 8; i++ {
		newval <<= 1
		newval += (val & 1)
		val >>= 1
	}
	return newval
}

// fixDesKey will make sure that exactly 8 bytes is used either by truncating or padding with nulls
// The bytes are then bit mirrored and returned
func fixDesKey(key []byte) []byte {
	tmp := key
	buf := make([]byte, 8)
	if len(tmp) <= 8 {
		copy(buf, tmp)
	} else {
		copy(buf, tmp[:8])
	}
	for i := 0; i < 8; i++ {
		buf[i] = fixDesKeyByte(buf[i])
	}
	return buf
}

func (fc *RFBConn) recvVNCChallenge(s []byte, err error) {
	glog.Debug("RFBConn recvVNCChallenge", hex.EncodeToString(s), len(s), err)
	key := core.Random(8)
	bk, err := des.NewCipher(fixDesKey(key))
	if err != nil {
		log.Printf("Error generating authentication cipher: %s\n", err.Error())
		return
	}
	result := make([]byte, 16)
	bk.Encrypt(result, s) //Encrypt first 8 bytes
	bk.Encrypt(result[8:], s[8:])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(result))
	fc.Write(result)
	core.StartReadBytes(4, fc, fc.recvSecurityResult)
}
func (fc *RFBConn) recvSecurityResult(s []byte, err error) {
	r := bytes.NewReader(s)
	result, _ := core.ReadUInt32BE(r)
	glog.Debug("RFBConn recvSecurityResult", result, err)
	if result == 1 {
		fc.Emit("error", fmt.Errorf("%s", "Authentification failed"))
		return
	}
	buff := &bytes.Buffer{}
	core.WriteUInt8(0, buff)
	fc.Write(buff.Bytes())
	core.StartReadBytes(20, fc, fc.recvServerInit)
}

type ServerInit struct {
	Width       uint16      `struc:"little"`
	Height      uint16      `struc:"little"`
	PixelFormat PixelFormat `struc:"little"`
}

func (fc *RFBConn) recvServerInit(s []byte, err error) {
	glog.Debug("RFBConn recvServerInit", len(s), err)
	r := bytes.NewReader(s)
	serverInit := &ServerInit{}
	err = struc.Unpack(r, serverInit)
	if err != nil {
		fc.Emit("error", err)
		return
	}
	glog.Infof("serverInit:%+v", serverInit)
	fc.s = serverInit
	fc.BitRect.Pf = &serverInit.PixelFormat
	core.StartReadBytes(4, fc, fc.checkServerName)
}
func (fc *RFBConn) checkServerName(s []byte, err error) {
	r := bytes.NewReader(s)
	result, _ := core.ReadUInt32BE(r)
	glog.Debug("RFBConn recvSecurityList", result, err)

	core.StartReadBytes(int(result), fc, fc.recvServerName)
}
func (fc *RFBConn) recvServerName(s []byte, err error) {
	glog.Debug("RFBConn recvServerName", string(s), err)
	//fc.sendPixelFormat()
	fc.sendSetEncoding()
	fc.sendFramebufferUpdateRequest(0, 0, 0, fc.s.Width, fc.s.Height)

	fc.Emit("ready")
	core.StartReadBytes(1, fc, fc.recvServerOrder)
}

func (fc *RFBConn) sendPixelFormat() {
	buff := &bytes.Buffer{}
	core.WriteUInt8(0, buff)
	core.WriteUInt16BE(0, buff)
	core.WriteUInt8(0, buff)
	err := struc.Pack(buff, fc.s.PixelFormat)
	if err != nil {
		fc.Emit("error", err)
		return
	}

	fc.Write(buff.Bytes())
}
func (fc *RFBConn) sendSetEncoding() {
	buff := &bytes.Buffer{}
	core.WriteUInt8(2, buff)
	core.WriteUInt8(0, buff)
	core.WriteUInt16BE(1, buff)
	core.WriteUInt32BE(0, buff)
	fc.Write(buff.Bytes())
}

type FrameBufferUpdateRequest struct {
	Incremental uint8
	X           uint16
	Y           uint16
	Width       uint16
	Height      uint16
}

func (fc *RFBConn) sendFramebufferUpdateRequest(Incremental uint8,
	X uint16,
	Y uint16,
	Width uint16,
	Height uint16) {
	buff := &bytes.Buffer{}
	core.WriteUInt8(3, buff)
	core.WriteUInt8(Incremental, buff)
	core.WriteUInt16BE(X, buff)
	core.WriteUInt16BE(Y, buff)
	core.WriteUInt16BE(Width, buff)
	core.WriteUInt16BE(Height, buff)
	fc.Write(buff.Bytes())
}
func (fc *RFBConn) recvServerOrder(s []byte, err error) {
	glog.Debug("RFBConn recvServerOrder", hex.EncodeToString(s), err)
	r := bytes.NewReader(s)
	packetType, _ := core.ReadUInt8(r)
	switch packetType {
	case 0:
		core.StartReadBytes(3, fc, fc.recvFrameBufferUpdateHeader)
	case 2:
		//core.StartReadBytes(4, fc, fc.recvServerName)
	case 3:
		core.StartReadBytes(7, fc, fc.recvServerCutTextHeader)
	default:
		glog.Errorf("Unknown message type %s", packetType)
	}

}
func (fc *RFBConn) recvFrameBufferUpdateHeader(s []byte, err error) {
	glog.Debug("RFBConn recvFrameBufferUpdateHeader", hex.EncodeToString(s), err)
	r := bytes.NewReader(s)
	core.ReadUInt8(r)
	NbRect, _ := core.ReadUint16BE(r)
	fc.NbRect = NbRect
	glog.Info("NbRect:", NbRect)
	core.StartReadBytes(12, fc, fc.recvRectHeader)
}
func (fc *RFBConn) recvRectHeader(s []byte, err error) {
	glog.Debug("RFBConn recvRectHeader", hex.EncodeToString(s), err)
	r := bytes.NewReader(s)
	rect := &Rectangle{}
	err = struc.Unpack(r, rect)
	if err != nil {
		fc.Emit("error", err)
		return
	}
	fc.BitRect.Rect = rect
	glog.Infof("rect:%+v, len=%d", rect, int(rect.Width)*int(rect.Height)*4)
	core.StartReadBytes(int(fc.s.Width)*int(fc.s.Height)*4, fc, fc.recvRectBody)
}
func (fc *RFBConn) recvRectBody(s []byte, err error) {
	glog.Debug("RFBConn recvRectBody", hex.EncodeToString(s), err)
	fc.BitRect.Data = s
	fc.Emit("update", fc.BitRect)
	fc.NbRect--
	glog.Info("fc.NbRect:", fc.NbRect)
	if fc.NbRect == 0 {
		fc.sendFramebufferUpdateRequest(1, 0, 0, fc.s.Width, fc.s.Height)
		core.StartReadBytes(1, fc, fc.recvServerOrder)
	} else {
		core.StartReadBytes(12, fc, fc.recvRectHeader)
	}
}

type ServerCutTextHeader struct {
	Padding [3]byte `struc:"little"`
	Size    uint32  `struc:"little"`
}

func (fc *RFBConn) recvServerCutTextHeader(s []byte, err error) {
	glog.Debug("RFBConn recvServerCutTextHeader", string(s), err)
	r := bytes.NewReader(s)
	header := &ServerCutTextHeader{}
	err = struc.Unpack(r, header)
	if err != nil {
		fc.Emit("error", err)
		return
	}

	core.StartReadBytes(int(header.Size), fc, fc.recvServerCutTextBody)
}
func (fc *RFBConn) recvServerCutTextBody(s []byte, err error) {
	glog.Debug("RFBConn recvServerCutTextBody", string(s), err)
	fc.Emit("CutText", s)
	core.StartReadBytes(1, fc, fc.recvServerOrder)
}

type PixelFormat struct {
	BitsPerPixel  uint8  `struc:"little"`
	Depth         uint8  `struc:"little"`
	BigEndianFlag bool   `struc:"little"`
	TrueColorFlag bool   `struc:"little"`
	RedMax        uint16 `struc:"little"`
	GreenMax      uint16 `struc:"little"`
	BlueMax       uint16 `struc:"little"`
	RedShift      uint8  `struc:"little"`
	GreenShift    uint8  `struc:"little"`
	BlueShift     uint8  `struc:"little"`
	padding       []byte `struc:"little"`
}

func NewPixelFormat() *PixelFormat {
	return &PixelFormat{
		32, 24, false, true, 255, 255, 255, 16, 8, 0, nil,
	}
}

type Rectangle struct {
	X        uint16 `struc:"little"`
	Y        uint16 `struc:"little"`
	Width    uint16 `struc:"little"`
	Height   uint16 `struc:"little"`
	Encoding uint32 `struc:"little"`
}
type RFB struct {
	core.Transport
	Version       string
	SecurityLevel uint8
	ServerName    string
	PixelFormat   *PixelFormat
	NbRect        int
	CurrentRect   *Rectangle
	Password      string
}

func NewRFB(t core.Transport) *RFB {
	fb := &RFB{t, RFB003008, SEC_INVALID, "", NewPixelFormat(), 0, &Rectangle{}, ""}

	fb.Once("data", fb.recvProtocolVersion)

	return fb
}

func (fb *RFB) recvProtocolVersion(version string) {
	if version != RFB003003 || version != RFB003007 || version != RFB003008 {
		version = RFB003008
	}
	glog.Infof("version:%s", version)
	buff := &bytes.Buffer{}
	buff.WriteString(version)
	fb.Write(buff.Bytes())
}
