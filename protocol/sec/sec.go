package sec

import (
	"bytes"
	"encoding/hex"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/emission"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/t125"
	"github.com/icodeface/grdp/protocol/t125/gcc"
	"unicode/utf16"
)

/**
 * SecurityFlag
 * @see http://msdn.microsoft.com/en-us/library/cc240579.aspx
 */
const (
	EXCHANGE_PKT       uint16 = 0x0001
	TRANSPORT_REQ             = 0x0002
	TRANSPORT_RSP             = 0x0004
	ENCRYPT                   = 0x0008
	RESET_SEQNO               = 0x0010
	IGNORE_SEQNO              = 0x0020
	INFO_PKT                  = 0x0040
	LICENSE_PKT               = 0x0080
	LICENSE_ENCRYPT_CS        = 0x0200
	LICENSE_ENCRYPT_SC        = 0x0200
	REDIRECTION_PKT           = 0x0400
	SECURE_CHECKSUM           = 0x0800
	AUTODETECT_REQ            = 0x1000
	AUTODETECT_RSP            = 0x2000
	HEARTBEAT                 = 0x4000
	FLAGSHI_VALID             = 0x8000
)

const (
	INFO_MOUSE                  uint32 = 0x00000001
	INFO_DISABLECTRLALTDEL             = 0x00000002
	INFO_AUTOLOGON                     = 0x00000008
	INFO_UNICODE                       = 0x00000010
	INFO_MAXIMIZESHELL                 = 0x00000020
	INFO_LOGONNOTIFY                   = 0x00000040
	INFO_COMPRESSION                   = 0x00000080
	INFO_ENABLEWINDOWSKEY              = 0x00000100
	INFO_REMOTECONSOLEAUDIO            = 0x00002000
	INFO_FORCE_ENCRYPTED_CS_PDU        = 0x00004000
	INFO_RAIL                          = 0x00008000
	INFO_LOGONERRORS                   = 0x00010000
	INFO_MOUSE_HAS_WHEEL               = 0x00020000
	INFO_PASSWORD_IS_SC_PIN            = 0x00040000
	INFO_NOAUDIOPLAYBACK               = 0x00080000
	INFO_USING_SAVED_CREDS             = 0x00100000
	INFO_AUDIOCAPTURE                  = 0x00200000
	INFO_VIDEO_DISABLE                 = 0x00400000
	INFO_CompressionTypeMask           = 0x00001E00
)

const (
	AF_INET  uint16 = 0x00002
	AF_INET6        = 0x0017
)

const (
	PERF_DISABLE_WALLPAPER          uint32 = 0x00000001
	PERF_DISABLE_FULLWINDOWDRAG            = 0x00000002
	PERF_DISABLE_MENUANIMATIONS            = 0x00000004
	PERF_DISABLE_THEMING                   = 0x00000008
	PERF_DISABLE_CURSOR_SHADOW             = 0x00000020
	PERF_DISABLE_CURSORSETTINGS            = 0x00000040
	PERF_ENABLE_FONT_SMOOTHING             = 0x00000080
	PERF_ENABLE_DESKTOP_COMPOSITION        = 0x00000100
)

type RDPExtendedInfo struct {
	clientAddressFamily uint16
	cbClientAddress     uint16
	clientAddress       []byte
	cbClientDir         uint16
	clientDir           []byte
	clientTimeZone      []byte
	clientSessionId     uint32
	performanceFlags    uint32
}

func NewExtendedInfo() *RDPExtendedInfo {
	return &RDPExtendedInfo{
		clientAddress:  []byte{0, 0},
		clientDir:      []byte{0, 0},
		clientTimeZone: make([]byte, 172),
	}
}

func (ext *RDPExtendedInfo) Serialize() []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt16LE(AF_INET, buff)                        // 0200
	core.WriteUInt16LE(uint16(len(ext.clientAddress)), buff) // 0200
	core.WriteBytes(ext.clientAddress, buff)                 // 0000
	core.WriteUInt16LE(uint16(len(ext.clientDir)), buff)     // 0200
	core.WriteBytes(ext.clientDir, buff)                     // 0000
	core.WriteBytes(ext.clientTimeZone, buff)
	core.WriteUInt32LE(ext.clientSessionId, buff)  // 00000000
	core.WriteUInt32LE(ext.performanceFlags, buff) // 00000000
	return buff.Bytes()
}

// 02000200000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

type RDPInfo struct {
	codePage         uint32
	flag             uint32
	cbDomain         uint16
	cbUserName       uint16
	cbPassword       uint16
	cbAlternateShell uint16
	cbWorkingDir     uint16
	domain           []byte
	userName         []byte
	password         []byte
	alternateShell   []byte
	workingDir       []byte
	extendedInfo     *RDPExtendedInfo
}

func NewRDPInfo() *RDPInfo {
	info := &RDPInfo{
		domain:         []byte{0, 0},
		userName:       []byte{0, 0},
		password:       []byte{0, 0},
		alternateShell: []byte{0, 0},
		workingDir:     []byte{0, 0},
		extendedInfo:   NewExtendedInfo(),
	}
	return info
}

func (o *RDPInfo) Serialize(hasExtended bool) []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt32LE(o.codePage, buff) // 0000000
	// 0530101
	core.WriteUInt32LE(INFO_MOUSE|INFO_UNICODE|INFO_LOGONNOTIFY|INFO_LOGONERRORS|INFO_DISABLECTRLALTDEL|INFO_ENABLEWINDOWSKEY, buff)
	core.WriteUInt16LE(uint16(len(o.domain)-2), buff)         // 001c
	core.WriteUInt16LE(uint16(len(o.userName)-2), buff)       // 0008
	core.WriteUInt16LE(uint16(len(o.password)-2), buff)       //000c
	core.WriteUInt16LE(uint16(len(o.alternateShell)-2), buff) //0000
	core.WriteUInt16LE(uint16(len(o.workingDir)-2), buff)     //0000
	core.WriteBytes(o.domain, buff)
	core.WriteBytes(o.userName, buff)
	core.WriteBytes(o.password, buff)
	core.WriteBytes(o.alternateShell, buff)
	core.WriteBytes(o.workingDir, buff)
	if hasExtended {
		core.WriteBytes(o.extendedInfo.Serialize(), buff)
	}

	return buff.Bytes()
}

type SEC struct {
	emission.Emitter
	transport   core.Transport
	info        *RDPInfo
	machineName string
	clientData  []interface{}
	serverData  []interface{}
}

type SecurityHeader struct {
	securityFlag   uint16
	securityFlagHi uint16
}

func NewSEC(t core.Transport) *SEC {
	sec := &SEC{
		*emission.NewEmitter(),
		t,
		NewRDPInfo(),
		"",
		nil,
		nil,
	}

	t.On("close", func() {
		sec.Emit("close")
	}).On("error", func(err error) {
		sec.Emit("error", err)
	})
	return sec
}

func (s *SEC) sendFlagged(flag uint16, data []byte) {
	glog.Debug("sendFlagged", hex.EncodeToString(data))
	buff := &bytes.Buffer{}
	core.WriteUInt16LE(flag, buff)
	core.WriteUInt16LE(0, buff)
	core.WriteBytes(data, buff)
	s.transport.Write(buff.Bytes())
}

type Client struct {
	*SEC
	userId    uint16
	channelId uint16
}

func NewClient(t core.Transport) *Client {
	c := &Client{
		SEC: NewSEC(t),
	}
	t.On("connect", c.connect)
	return c
}

func (c *Client) SetUser(user string) {
	// jack => 6a00610063006b000000
	buff := &bytes.Buffer{}
	for _, ch := range utf16.Encode([]rune(user)) {
		core.WriteUInt16LE(ch, buff)
	}
	core.WriteUInt16LE(0, buff)
	c.info.userName = buff.Bytes()
}

func (c *Client) SetPwd(pwd string) {
	buff := &bytes.Buffer{}
	for _, ch := range utf16.Encode([]rune(pwd)) {
		core.WriteUInt16LE(ch, buff)
	}
	core.WriteUInt16LE(0, buff)
	c.info.password = buff.Bytes()
}

func (c *Client) SetDomain(domain string) {
	buff := &bytes.Buffer{}
	for _, ch := range utf16.Encode([]rune(domain)) {
		core.WriteUInt16LE(ch, buff)
	}
	core.WriteUInt16LE(0, buff)
	c.info.domain = buff.Bytes()
}

func (c *Client) connect(clientData []interface{}, serverData []interface{}, userId uint16, channels []t125.MCSChannelInfo) {
	glog.Debug("sec on connect")
	// todo
	c.clientData = clientData
	c.serverData = serverData
	c.userId = userId
	c.channelId = 0
	c.sendInfoPkt()
}

func (c *Client) sendInfoPkt() {
	glog.Debug("sendInfoPkt", hex.EncodeToString(c.info.Serialize(true)))
	c.sendFlagged(INFO_PKT, c.info.Serialize(c.clientData[0].(*gcc.ClientCoreData).RdpVersion == gcc.RDP_VERSION_5_PLUS))
}
