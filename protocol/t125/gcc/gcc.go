package gcc

import (
	"bytes"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/t125/per"
)

var t124_02_98_oid = []byte{0, 0, 20, 124, 0, 1}
var h221_cs_key = "Duca"
var h221_sc_key = "McDn"

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240509.aspx
 */
type Message uint16

const (
	//server -> client
	SC_CORE     Message = 0x0C01
	SC_SECURITY         = 0x0C02
	SC_NET              = 0x0C03
	//client -> server
	CS_CORE     = 0xC001
	CS_SECURITY = 0xC002
	CS_NET      = 0xC003
	CS_CLUSTER  = 0xC004
	CS_MONITOR  = 0xC005
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type ColorDepth uint16

const (
	RNS_UD_COLOR_8BPP      ColorDepth = 0xCA01
	RNS_UD_COLOR_16BPP_555            = 0xCA02
	RNS_UD_COLOR_16BPP_565            = 0xCA03
	RNS_UD_COLOR_24BPP                = 0xCA04
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type HighColor uint16

const (
	HIGH_COLOR_4BPP  HighColor = 0x0004
	HIGH_COLOR_8BPP            = 0x0008
	HIGH_COLOR_15BPP           = 0x000f
	HIGH_COLOR_16BPP           = 0x0010
	HIGH_COLOR_24BPP           = 0x0018
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type Support uint16

const (
	RNS_UD_24BPP_SUPPORT uint16 = 0x0001
	RNS_UD_16BPP_SUPPORT        = 0x0002
	RNS_UD_15BPP_SUPPORT        = 0x0004
	RNS_UD_32BPP_SUPPORT        = 0x0008
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type CapabilityFlag uint16

const (
	RNS_UD_CS_SUPPORT_ERRINFO_PDU        uint16 = 0x0001
	RNS_UD_CS_WANT_32BPP_SESSION                = 0x0002
	RNS_UD_CS_SUPPORT_STATUSINFO_PDU            = 0x0004
	RNS_UD_CS_STRONG_ASYMMETRIC_KEYS            = 0x0008
	RNS_UD_CS_UNUSED                            = 0x0010
	RNS_UD_CS_VALID_CONNECTION_TYPE             = 0x0020
	RNS_UD_CS_SUPPORT_MONITOR_LAYOUT_PDU        = 0x0040
	RNS_UD_CS_SUPPORT_NETCHAR_AUTODETECT        = 0x0080
	RNS_UD_CS_SUPPORT_DYNVC_GFX_PROTOCOL        = 0x0100
	RNS_UD_CS_SUPPORT_DYNAMIC_TIME_ZONE         = 0x0200
	RNS_UD_CS_SUPPORT_HEARTBEAT_PDU             = 0x0400
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type ConnectionType uint8

const (
	CONNECTION_TYPE_MODEM          ConnectionType = 0x01
	CONNECTION_TYPE_BROADBAND_LOW                 = 0x02
	CONNECTION_TYPE_SATELLITEV                    = 0x03
	CONNECTION_TYPE_BROADBAND_HIGH                = 0x04
	CONNECTION_TYPE_WAN                           = 0x05
	CONNECTION_TYPE_LAN                           = 0x06
	CONNECTION_TYPE_AUTODETECT                    = 0x07
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240510.aspx
 */
type VERSION uint32

const (
	RDP_VERSION_4      VERSION = 0x00080001
	RDP_VERSION_5_PLUS         = 0x00080004
)

type Sequence uint16

const (
	RNS_UD_SAS_DEL Sequence = 0xAA03
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240511.aspx
 */
type EncryptionMethod uint32

const (
	ENCRYPTION_FLAG_40BIT  uint32 = 0x00000001
	ENCRYPTION_FLAG_128BIT        = 0x00000002
	ENCRYPTION_FLAG_56BIT         = 0x00000008
	FIPS_ENCRYPTION_FLAG          = 0x00000010
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240518.aspx
 */
type EncryptionLevel uint32

const (
	ENCRYPTION_LEVEL_NONE              EncryptionLevel = 0x00000000
	ENCRYPTION_LEVEL_LOW                               = 0x00000001
	ENCRYPTION_LEVEL_CLIENT_COMPATIBLE                 = 0x00000002
	ENCRYPTION_LEVEL_HIGH                              = 0x00000003
	ENCRYPTION_LEVEL_FIPS                              = 0x00000004
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240513.aspx
 */
type ChannelOptions uint32

const (
	CHANNEL_OPTION_INITIALIZED   ChannelOptions = 0x80000000
	CHANNEL_OPTION_ENCRYPT_RDP                  = 0x40000000
	CHANNEL_OPTION_ENCRYPT_SC                   = 0x20000000
	CHANNEL_OPTION_ENCRYPT_CS                   = 0x10000000
	CHANNEL_OPTION_PRI_HIGH                     = 0x08000000
	CHANNEL_OPTION_PRI_MED                      = 0x04000000
	CHANNEL_OPTION_PRI_LOW                      = 0x02000000
	CHANNEL_OPTION_COMPRESS_RDP                 = 0x00800000
	CHANNEL_OPTION_COMPRESS                     = 0x00400000
	CHANNEL_OPTION_SHOW_PROTOCOL                = 0x00200000
	REMOTE_CONTROL_PERSISTENT                   = 0x00100000
)

/**
 * IBM_101_102_KEYS is the most common keyboard type
 */
type KeyboardType uint32

const (
	KT_IBM_PC_XT_83_KEY KeyboardType = 0x00000001
	KT_OLIVETTI                      = 0x00000002
	KT_IBM_PC_AT_84_KEY              = 0x00000003
	KT_IBM_101_102_KEYS              = 0x00000004
	KT_NOKIA_1050                    = 0x00000005
	KT_NOKIA_9140                    = 0x00000006
	KT_JAPANESE                      = 0x00000007
)

/**
 * @see http://technet.microsoft.com/en-us/library/cc766503%28WS.10%29.aspx
 */
type KeyboardLayout uint32

const (
	ARABIC              KeyboardLayout = 0x00000401
	BULGARIAN                          = 0x00000402
	CHINESE_US_KEYBOARD                = 0x00000404
	CZECH                              = 0x00000405
	DANISH                             = 0x00000406
	GERMAN                             = 0x00000407
	GREEK                              = 0x00000408
	US                                 = 0x00000409
	SPANISH                            = 0x0000040a
	FINNISH                            = 0x0000040b
	FRENCH                             = 0x0000040c
	HEBREW                             = 0x0000040d
	HUNGARIAN                          = 0x0000040e
	ICELANDIC                          = 0x0000040f
	ITALIAN                            = 0x00000410
	JAPANESE                           = 0x00000411
	KOREAN                             = 0x00000412
	DUTCH                              = 0x00000413
	NORWEGIAN                          = 0x00000414
)

/**
 * @see http://msdn.microsoft.com/en-us/library/cc240521.aspx
 */
type CertificateType uint32

const (
	CERT_CHAIN_VERSION_1 CertificateType = 0x00000001
	CERT_CHAIN_VERSION_2                 = 0x00000002
)

type ChannelDef struct {
	Name    [8]byte
	Options uint32
}

type ClientCoreData struct {
	RdpVersion             VERSION
	DesktopWidth           uint16
	DesktopHeight          uint16
	ColorDepth             ColorDepth
	SasSequence            Sequence
	KbdLayout              KeyboardLayout
	ClientBuild            uint32
	ClientName             [32]byte
	KeyboardType           uint32
	KeyboardSubType        uint32
	KeyboardFnKeys         uint32
	ImeFileName            [64]byte
	PostBeta2ColorDepth    ColorDepth //optional
	ClientProductId        uint16     //optional
	SerialNumber           uint32     //optional
	HighColorDepth         HighColor  //optional
	SupportedColorDepths   uint16     //optional
	EarlyCapabilityFlags   uint16     //optional
	ClientDigProductId     [64]byte   //optional
	ConnectionType         uint8      //optional
	Pad1octet              uint8      //optional
	ServerSelectedProtocol uint32     //optional
}

func NewClientCoreData() *ClientCoreData {
	return &ClientCoreData{
		RDP_VERSION_5_PLUS, 1280, 800, RNS_UD_COLOR_8BPP,
		RNS_UD_SAS_DEL, US, 3790, [32]byte{'m', 's', 't', 's', 'c'}, KT_IBM_101_102_KEYS,
		0, 12, [64]byte{}, RNS_UD_COLOR_8BPP, 1, 0, HIGH_COLOR_24BPP,
		RNS_UD_15BPP_SUPPORT | RNS_UD_16BPP_SUPPORT | RNS_UD_24BPP_SUPPORT | RNS_UD_32BPP_SUPPORT,
		RNS_UD_CS_SUPPORT_ERRINFO_PDU, [64]byte{}, 0, 0, 0}
}

func (data *ClientCoreData) Block() []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt16LE(CS_CORE, buff)                  // 01C0
	core.WriteUInt16LE(0xd8, buff)                     // d8
	core.WriteUInt32LE(uint32(data.RdpVersion), buff)  // 00040008
	core.WriteUInt16LE(data.DesktopWidth, buff)        // 0000
	core.WriteUInt16LE(data.DesktopHeight, buff)       // 5200
	core.WriteUInt16LE(uint16(data.ColorDepth), buff)  // 301c
	core.WriteUInt16LE(uint16(data.SasSequence), buff) // a03a
	core.WriteUInt32LE(uint32(data.KbdLayout), buff)   // a0904000
	core.WriteUInt32LE(data.ClientBuild, buff)         // 0ce0e000
	core.WriteBytes(data.ClientName[:], buff)          //
	core.WriteUInt32LE(data.KeyboardType, buff)        //    00000004
	core.WriteUInt32LE(data.KeyboardSubType, buff)     // 00000000
	core.WriteUInt32LE(data.KeyboardFnKeys, buff)      //
	core.WriteBytes(data.ImeFileName[:], buff)
	core.WriteUInt16LE(uint16(data.PostBeta2ColorDepth), buff)
	core.WriteUInt16LE(data.ClientProductId, buff)
	core.WriteUInt32LE(data.SerialNumber, buff)
	core.WriteUInt16LE(uint16(data.HighColorDepth), buff)
	core.WriteUInt16LE(data.SupportedColorDepths, buff)
	core.WriteUInt16LE(data.EarlyCapabilityFlags, buff)
	core.WriteBytes(data.ClientDigProductId[:], buff)
	core.WriteUInt8(data.ConnectionType, buff)
	core.WriteUInt8(data.Pad1octet, buff)
	core.WriteUInt32LE(data.ServerSelectedProtocol, buff)
	return buff.Bytes()
}

// grdp:
// rdpy: 01c0d800040008000005200301ca03aa09040000ce0e0000720064007000790000000000000000000000000000000000000000000000000004000000000000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001ca01000000000018000f00010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000

type ClientNetworkData struct {
	ChannelCount    uint32
	ChannelDefArray []ChannelDef
}

func NewClientNetworkData() *ClientNetworkData {
	return &ClientNetworkData{}
}

func (d *ClientNetworkData) Block() []byte {
	// 03c0080000000000
	buff := &bytes.Buffer{}
	core.WriteUInt16LE(CS_NET, buff) // type
	core.WriteUInt16LE(0x08, buff)   // len 8
	buff.Write([]byte{0, 0, 0, 0})   // data
	return buff.Bytes()
}

type ClientSecurityData struct {
	EncryptionMethods    uint32
	ExtEncryptionMethods uint32
}

func NewClientSecurityData() *ClientSecurityData {
	return &ClientSecurityData{
		ENCRYPTION_FLAG_40BIT | ENCRYPTION_FLAG_56BIT | ENCRYPTION_FLAG_128BIT,
		00}
}

func (d *ClientSecurityData) Block() []byte {
	// 02c0 0c000b00000000000000
	buff := &bytes.Buffer{}
	core.WriteUInt16LE(CS_SECURITY, buff) // type
	core.WriteUInt16LE(0x0c, buff)        // len 12
	core.WriteUInt32LE(d.EncryptionMethods, buff)
	core.WriteUInt32LE(d.ExtEncryptionMethods, buff)
	return buff.Bytes()
}

type ServerCoreData struct {
	RdpVersion              VERSION
	ClientRequestedProtocol uint32 //optional
	EarlyCapabilityFlags    uint32 //optional
	raw                     []byte
}

func NewServerCoreData() *ServerCoreData {
	return &ServerCoreData{
		RDP_VERSION_5_PLUS, 0, 0, []byte{}}
}

func (d *ServerCoreData) Serialize() []byte {
	return []byte{}
}

type ServerNetworkData struct {
	ChannelIdArray []uint16
}

func NewServerNetworkData() *ServerNetworkData {
	return &ServerNetworkData{}
}

type ServerSecurityData struct {
	EncryptionMethod uint32
	EncryptionLevel  uint32
	raw              []byte
}

func NewServerSecurityData() *ServerSecurityData {
	return &ServerSecurityData{
		0, 0, []byte{}}
}

func MakeConferenceCreateRequest(userData []byte) []byte {
	buff := &bytes.Buffer{}
	per.WriteChoice(0, buff)                        // 00
	per.WriteObjectIdentifier(t124_02_98_oid, buff) // 05:00:14:7c:00:01
	per.WriteLength(len(userData)+14, buff)
	per.WriteChoice(0, buff)                   // 00
	per.WriteSelection(0x08, buff)             // 08
	per.WriteNumericString("1", 1, buff)       // 00 10
	per.WritePadding(1, buff)                  // 00
	per.WriteNumberOfSet(1, buff)              // 01
	per.WriteChoice(0xc0, buff)                // c0
	per.WriteOctetStream(h221_cs_key, 4, buff) // 00 44:75:63:61
	per.WriteOctetStream(string(userData), 0, buff)
	return buff.Bytes()
}

// rdpy: 000500147c000180fa000800100001c00044756361
// userData below
// 80ec01c0d800040008000005200301ca03aa09040000ce0e0000720064007000790000000000000000000000000000000000000000000000000004000000000000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001ca01000000000018000f0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000003c008000000000002c00c000b00000000000000

// grdp: 000500147c000180f6000800100001c00044756361
// userData below
// 80e801c0d800040008000005200301ca03aa09040000ce0e00006d73747363000000000000000000000000000000000000000000000000000000040000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001ca01000000000018000f0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000003c008000000000002c00c000b00000000000000

func ReadConferenceCreateResponse(data []byte) []interface{} {
	// todo
	glog.Debug("ReadConferenceCreateResponse todo")
	ret := make([]interface{}, 0)
	return ret
}
