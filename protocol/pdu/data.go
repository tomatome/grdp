package pdu

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
	"github.com/lunixbochs/struc"
	"io"
)

const (
	PDUTYPE_DEMANDACTIVEPDU  uint16 = 0x11
	PDUTYPE_CONFIRMACTIVEPDU        = 0x13
	PDUTYPE_DEACTIVATEALLPDU        = 0x16
	PDUTYPE_DATAPDU                 = 0x17
	PDUTYPE_SERVER_REDIR_PKT        = 0x1A
)

type ShareDataHeader struct {
	SharedId           uint32 `struc:"little"`
	Padding1           uint8  `struc:"little"`
	StreamId           uint8  `struc:"little"`
	UncompressedLength uint16 `struc:"little"`
	PDUType2           uint8  `struc:"little"`
	CompressedType     uint8  `struc:"little"`
	CompressedLength   uint16 `struc:"little"`
}

type ShareControlHeader struct {
	TotalLength uint16 `struc:"little"`
	PDUType     uint16 `struc:"little"`
	PDUSource   uint16 `struc:"little"`
}

type PDUMessage interface {
	Type() uint16
	Serialize() []byte
}

type DemandActivePDU struct {
	SharedId                   uint32       `struc:"little"`
	LengthSourceDescriptor     uint16       `struc:"little,sizeof=SourceDescriptor"`
	LengthCombinedCapabilities uint16       `struc:"little"`
	SourceDescriptor           string       `struc:"sizefrom=LengthSourceDescriptor"`
	NumberCapabilities         uint16       `struc:"little,sizeof=CapabilitySets"`
	Pad2Octets                 uint16       `struc:"little"`
	CapabilitySets             []Capability `struc:"sizefrom=NumberCapabilities"`
	SessionId                  uint32       `struc:"little"`
}

func (d *DemandActivePDU) Type() uint16 {
	return PDUTYPE_DEMANDACTIVEPDU
}

func (d *DemandActivePDU) Serialize() []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt32LE(d.SharedId, buff)
	core.WriteUInt16LE(d.LengthSourceDescriptor, buff)
	core.WriteUInt16LE(d.LengthCombinedCapabilities, buff)
	core.WriteBytes([]byte(d.SourceDescriptor), buff)
	core.WriteUInt16LE(uint16(len(d.CapabilitySets)), buff)
	core.WriteUInt16LE(d.Pad2Octets, buff)
	for _, cap := range d.CapabilitySets {
		core.WriteUInt16LE(uint16(cap.Type()), buff)
		capBuff := &bytes.Buffer{}
		struc.Pack(capBuff, cap)
		capBytes := capBuff.Bytes()
		core.WriteUInt16LE(uint16(len(capBytes)+4), buff)
		core.WriteBytes(capBytes, buff)
	}
	core.WriteUInt32LE(d.SessionId, buff)
	return buff.Bytes()
}

func readDemandActivePDU(r io.Reader) (*DemandActivePDU, error) {
	d := &DemandActivePDU{}
	var err error
	d.SharedId, err = core.ReadUInt32LE(r)
	if err != nil {
		return nil, err
	}
	d.LengthSourceDescriptor, err = core.ReadUint16LE(r)
	d.LengthCombinedCapabilities, err = core.ReadUint16LE(r)
	sourceDescriptorBytes, err := core.ReadBytes(int(d.LengthSourceDescriptor), r)
	if err != nil {
		return nil, err
	}
	d.SourceDescriptor = string(sourceDescriptorBytes)
	d.NumberCapabilities, err = core.ReadUint16LE(r)
	d.Pad2Octets, err = core.ReadUint16LE(r)
	d.CapabilitySets = make([]Capability, 0)
	glog.Debug("NumberCapabilities is", d.NumberCapabilities)
	for i := 0; i < int(d.NumberCapabilities); i++ {
		capType, err := core.ReadUint16LE(r)
		if err != nil {
			return nil, err
		}
		capLen, err := core.ReadUint16LE(r)
		if err != nil {
			return nil, err
		}
		capBytes, err := core.ReadBytes(int(capLen)-4, r)
		if err != nil {
			return nil, err
		}
		capReader := bytes.NewReader(capBytes)

		var c Capability
		switch CapsType(capType) {
		case CAPSTYPE_GENERAL:
			c = &GeneralCapability{}
		case CAPSTYPE_BITMAP:
			c = &BitmapCapability{}
		case CAPSTYPE_ORDER:
			c = &OrderCapability{}
		case CAPSTYPE_BITMAPCACHE:
			c = &BitmapCacheCapability{}
		case CAPSTYPE_POINTER:
			c = &PointerCapability{}
		case CAPSTYPE_INPUT:
			c = &InputCapability{}
		case CAPSTYPE_BRUSH:
			c = &BrushCapability{}
		case CAPSTYPE_GLYPHCACHE:
			c = &GlyphCapability{}
		case CAPSTYPE_OFFSCREENCACHE:
			c = &OffscreenBitmapCacheCapability{}
		case CAPSTYPE_VIRTUALCHANNEL:
			c = &VirtualChannelCapability{}
		case CAPSTYPE_SOUND:
			c = &SoundCapability{}
		case CAPSTYPE_CONTROL:
			c = &ControlCapability{}
		case CAPSTYPE_ACTIVATION:
			c = &WindowActivationCapability{}
		case CAPSTYPE_FONT:
			c = &FontCapability{}
		case CAPSTYPE_COLORCACHE:
			c = &ColorCacheCapability{}
		case CAPSTYPE_SHARE:
			c = &ShareCapability{}
		case CAPSETTYPE_MULTIFRAGMENTUPDATE:
			c = &MultiFragmentUpdate{}
		case CAPSTYPE_DRAWGDIPLUS:
			c = &DrawGDIPlusCapability{}
		case CAPSETTYPE_BITMAP_CODECS:
			c = &BitmapCodecsCapability{}
		case CAPSTYPE_BITMAPCACHE_HOSTSUPPORT:
			c = &BitmapCacheHostSupportCapability{}
		case CAPSETTYPE_LARGE_POINTER:
			c = &LargePointerCapability{}
		case CAPSTYPE_RAIL:
			c = &RemoteProgramsCapability{}
		case CAPSTYPE_WINDOW:
			c = &WindowListCapability{}
		case CAPSETTYPE_COMPDESK:
			c = &DesktopCompositionCapability{}
		case CAPSETTYPE_SURFACE_COMMANDS:
			c = &SurfaceCommandsCapability{}
		default:
			glog.Error("unknown Capability type", fmt.Sprintf("0x%04x", capType))
			c = nil
		}

		if c != nil {
			if err := struc.Unpack(capReader, c); err != nil {
				glog.Error("Capability unpack error", err, fmt.Sprintf("0x%04x", capType), hex.EncodeToString(capBytes))
				return nil, err
			}
			d.CapabilitySets = append(d.CapabilitySets, c)
		}
	}
	d.SessionId, err = core.ReadUInt32LE(r)
	if err != nil {
		return nil, err
	}
	return d, nil
}

type ConfirmActivePDU struct {
	SharedId                   uint32       `struc:"little"`
	OriginatorId               uint16       `struc:"little"`
	LengthSourceDescriptor     uint16       `struc:"little,sizeof=SourceDescriptor"`
	LengthCombinedCapabilities uint16       `struc:"little"`
	SourceDescriptor           string       `struc:"sizefrom=LengthSourceDescriptor"`
	NumberCapabilities         uint16       `struc:"little,sizeof=CapabilitySets"`
	Pad2Octets                 uint16       `struc:"pad"`
	CapabilitySets             []Capability `struc:"sizefrom=NumberCapabilities"`
}

func (*ConfirmActivePDU) Type() uint16 {
	return PDUTYPE_CONFIRMACTIVEPDU
}

func (c *ConfirmActivePDU) Serialize() []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt32LE(c.SharedId, buff)
	core.WriteUInt16LE(c.OriginatorId, buff)
	core.WriteUInt16LE(uint16(len(c.SourceDescriptor)), buff)

	capsBuff := &bytes.Buffer{}
	for _, capa := range c.CapabilitySets {
		struc.Pack(capsBuff, capa)
	}
	capsBytes := capsBuff.Bytes()

	// lengthCombinedCapabilities = UInt16Le(lambda:(sizeof(numberCapabilities) + sizeof(pad2Octets) + sizeof(capabilitySets)))
	core.WriteUInt16LE(uint16(2+2+len(capsBytes)), buff)

	core.WriteBytes([]byte(c.SourceDescriptor), buff)
	core.WriteUInt16LE(uint16(len(c.CapabilitySets)), buff)
	core.WriteUInt16LE(c.Pad2Octets, buff)
	core.WriteBytes(capsBytes, buff)
	return buff.Bytes()
}

func NewConfirmActivePDU() *ConfirmActivePDU {
	return &ConfirmActivePDU{
		OriginatorId:   0x03EA,
		CapabilitySets: make([]Capability, 0),
	}
}

type PDU struct {
	ShareCtrlHeader *ShareControlHeader
	Message         PDUMessage
}

func NewPDU(userId uint16, message PDUMessage) *PDU {
	pdu := &PDU{}
	pdu.ShareCtrlHeader = &ShareControlHeader{
		TotalLength: uint16(len(message.Serialize())),
		PDUType:     message.Type(),
		PDUSource:   userId,
	}
	pdu.Message = message
	return pdu
}

func readPDU(r io.Reader) (*PDU, error) {
	pdu := &PDU{}
	var err error
	header := &ShareControlHeader{}
	err = struc.Unpack(r, header)
	if err != nil {
		return nil, err
	}
	pdu.ShareCtrlHeader = header

	switch pdu.ShareCtrlHeader.PDUType {
	case PDUTYPE_DEMANDACTIVEPDU:
		d, err := readDemandActivePDU(r)
		if err != nil {
			return nil, err
		}
		pdu.Message = d
	default:
		glog.Error("PDU invalid pdu type")
	}
	return pdu, err
}

func (p *PDU) serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, p.ShareCtrlHeader)
	core.WriteBytes(p.Message.Serialize(), buff)
	return buff.Bytes()
}
