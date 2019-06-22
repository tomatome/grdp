package pdu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
	"github.com/lunixbochs/struc"
	"io"
)

const (
	PDUTYPE_DEMANDACTIVEPDU  = 0x11
	PDUTYPE_CONFIRMACTIVEPDU = 0x13
	PDUTYPE_DEACTIVATEALLPDU = 0x16
	PDUTYPE_DATAPDU          = 0x17
	PDUTYPE_SERVER_REDIR_PKT = 0x1A
)

const (
	PDUTYPE2_UPDATE                      = 0x02
	PDUTYPE2_CONTROL                     = 0x14
	PDUTYPE2_POINTER                     = 0x1B
	PDUTYPE2_INPUT                       = 0x1C
	PDUTYPE2_SYNCHRONIZE                 = 0x1F
	PDUTYPE2_REFRESH_RECT                = 0x21
	PDUTYPE2_PLAY_SOUND                  = 0x22
	PDUTYPE2_SUPPRESS_OUTPUT             = 0x23
	PDUTYPE2_SHUTDOWN_REQUEST            = 0x24
	PDUTYPE2_SHUTDOWN_DENIED             = 0x25
	PDUTYPE2_SAVE_SESSION_INFO           = 0x26
	PDUTYPE2_FONTLIST                    = 0x27
	PDUTYPE2_FONTMAP                     = 0x28
	PDUTYPE2_SET_KEYBOARD_INDICATORS     = 0x29
	PDUTYPE2_BITMAPCACHE_PERSISTENT_LIST = 0x2B
	PDUTYPE2_BITMAPCACHE_ERROR_PDU       = 0x2C
	PDUTYPE2_SET_KEYBOARD_IME_STATUS     = 0x2D
	PDUTYPE2_OFFSCRCACHE_ERROR_PDU       = 0x2E
	PDUTYPE2_SET_ERROR_INFO_PDU          = 0x2F
	PDUTYPE2_DRAWNINEGRID_ERROR_PDU      = 0x30
	PDUTYPE2_DRAWGDIPLUS_ERROR_PDU       = 0x31
	PDUTYPE2_ARC_STATUS_PDU              = 0x32
	PDUTYPE2_STATUS_INFO_PDU             = 0x36
	PDUTYPE2_MONITOR_LAYOUT_PDU          = 0x37
)

const (
	CTRLACTION_REQUEST_CONTROL = 0x0001
	CTRLACTION_GRANTED_CONTROL = 0x0002
	CTRLACTION_DETACH          = 0x0003
	CTRLACTION_COOPERATE       = 0x0004
)

const (
	STREAM_UNDEFINED = 0x00
	STREAM_LOW       = 0x01
	STREAM_MED       = 0x02
	STREAM_HI        = 0x04
)

const (
	FASTPATH_UPDATETYPE_ORDERS       = 0x0
	FASTPATH_UPDATETYPE_BITMAP       = 0x1
	FASTPATH_UPDATETYPE_PALETTE      = 0x2
	FASTPATH_UPDATETYPE_SYNCHRONIZE  = 0x3
	FASTPATH_UPDATETYPE_SURFCMDS     = 0x4
	FASTPATH_UPDATETYPE_PTR_NULL     = 0x5
	FASTPATH_UPDATETYPE_PTR_DEFAULT  = 0x6
	FASTPATH_UPDATETYPE_PTR_POSITION = 0x8
	FASTPATH_UPDATETYPE_COLOR        = 0x9
	FASTPATH_UPDATETYPE_CACHED       = 0xA
	FASTPATH_UPDATETYPE_POINTER      = 0xB
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

func NewShareDataHeader(size int, type2 uint8, shareId uint32) *ShareDataHeader {
	return &ShareDataHeader{
		SharedId:           shareId,
		PDUType2:           type2,
		StreamId:           STREAM_LOW,
		UncompressedLength: uint16(size + 4),
	}
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
		c, err := readCapability(r)
		if err != nil {
			return nil, err
		}
		d.CapabilitySets = append(d.CapabilitySets, c)
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
	Pad2Octets                 uint16       `struc:"little"`
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
		core.WriteUInt16LE(uint16(capa.Type()), capsBuff)
		capBuff := &bytes.Buffer{}
		struc.Pack(capBuff, capa)
		if capa.Type() == CAPSTYPE_INPUT {
			core.WriteBytes([]byte{0x0c, 0x00, 0x00, 0x00}, capBuff)
		}
		capBytes := capBuff.Bytes()
		core.WriteUInt16LE(uint16(len(capBytes)+4), capsBuff)
		core.WriteBytes(capBytes, capsBuff)
	}
	capsBytes := capsBuff.Bytes()

	core.WriteUInt16LE(uint16(2+2+len(capsBytes)), buff)
	core.WriteBytes([]byte(c.SourceDescriptor), buff)
	core.WriteUInt16LE(uint16(len(c.CapabilitySets)), buff)
	core.WriteUInt16LE(c.Pad2Octets, buff)
	core.WriteBytes(capsBytes, buff)
	return buff.Bytes()
}

// 9401 => share control header
// 1300 => share control header
// ec03 => share control header
// ea030100  => shareId 66538
// ea03 => OriginatorId
// 0400
// 8001 => LengthCombinedCapabilities
// 72647079
// 0c00 => NumberCapabilities 12
// 0000
// caps below
// 010018000100030000020000000015040000000000000000
// 02001c00180001000100010000052003000000000100000001000000
// 030058000000000000000000000000000000000000000000010014000000010000000a0000000000000000000000000000000000000000000000000000000000000000000000000000000000008403000000000000000000
// 04002800000000000000000000000000000000000000000000000000000000000000000000000000
// 0800080000001400
// 0c00080000000000
// 0d005c001500000009040000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c000000
// 0f00080000000000
// 10003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
// 11000c000000000000000000
// 14000c000000000000000000
// 1a00080000000000

func NewConfirmActivePDU() *ConfirmActivePDU {
	return &ConfirmActivePDU{
		OriginatorId:     0x03EA,
		CapabilitySets:   make([]Capability, 0),
		SourceDescriptor: "rdpy",
	}
}

func readConfirmActivePDU(r io.Reader) (*ConfirmActivePDU, error) {
	p := &ConfirmActivePDU{}
	var err error
	p.SharedId, err = core.ReadUInt32LE(r)
	if err != nil {
		return nil, err
	}
	p.OriginatorId, err = core.ReadUint16LE(r)
	p.LengthSourceDescriptor, err = core.ReadUint16LE(r)
	p.LengthCombinedCapabilities, err = core.ReadUint16LE(r)

	sourceDescriptorBytes, err := core.ReadBytes(int(p.LengthSourceDescriptor), r)
	if err != nil {
		return nil, err
	}
	p.SourceDescriptor = string(sourceDescriptorBytes)
	p.NumberCapabilities, err = core.ReadUint16LE(r)
	p.Pad2Octets, err = core.ReadUint16LE(r)

	p.CapabilitySets = make([]Capability, 0)
	for i := 0; i < int(p.NumberCapabilities); i++ {
		c, err := readCapability(r)
		if err != nil {
			return nil, err
		}
		p.CapabilitySets = append(p.CapabilitySets, c)
	}
	return p, nil
}

type DeactiveAllPDU struct {
	ShareId                uint32 `struc:"little"`
	LengthSourceDescriptor uint16 `struc:"little,sizeof=SourceDescriptor"`
	SourceDescriptor       []byte
}

func (*DeactiveAllPDU) Type() uint16 {
	return PDUTYPE_DEACTIVATEALLPDU
}

func (d *DeactiveAllPDU) Serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, d)
	return buff.Bytes()
}

func readDeactiveAllPDU(r io.Reader) (*DeactiveAllPDU, error) {
	p := &DeactiveAllPDU{}
	err := struc.Unpack(r, p)
	return p, err
}

type DataPDU struct {
	Header *ShareDataHeader
	Data   DataPDUData
}

func (*DataPDU) Type() uint16 {
	return PDUTYPE_DATAPDU
}

func (d *DataPDU) Serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, d.Header)
	struc.Pack(buff, d.Data)
	return buff.Bytes()
}

func NewDataPDU(data DataPDUData, shareId uint32) *DataPDU {
	dataBuff := &bytes.Buffer{}
	struc.Pack(dataBuff, data)
	return &DataPDU{
		Header: NewShareDataHeader(len(dataBuff.Bytes()), data.Type2(), shareId),
		Data:   data,
	}
}

func readDataPDU(r io.Reader) (*DataPDU, error) {
	header := &ShareDataHeader{}
	err := struc.Unpack(r, header)
	if err != nil {
		glog.Error("read data pdu header error", err)
		return nil, err
	}
	var d DataPDUData
	switch header.PDUType2 {
	case PDUTYPE2_SYNCHRONIZE:
		d = &SynchronizeDataPDU{}
	case PDUTYPE2_CONTROL:
		d = &ControlDataPDU{}
	case PDUTYPE2_FONTLIST:
		d = &FontListDataPDU{}
	case PDUTYPE2_SET_ERROR_INFO_PDU:
		d = &ErrorInfoDataPDU{}
	case PDUTYPE2_FONTMAP:
		d = &FontMapDataPDU{}
	default:
		err = errors.New(fmt.Sprintf("Unknown data pdu type2 0x%02x", header.PDUType2))
		glog.Error(err)
		return nil, err
	}
	err = struc.Unpack(r, d)
	if err != nil {
		glog.Error("read data pdu data error", err)
		return nil, err
	}
	p := &DataPDU{
		Header: header,
		Data:   d,
	}
	return p, nil
}

type DataPDUData interface {
	Type2() uint8
}

type SynchronizeDataPDU struct {
	MessageType uint16 `struc:"little"`
	TargetUser  uint16 `struc:"little"`
}

func (*SynchronizeDataPDU) Type2() uint8 {
	return PDUTYPE2_SYNCHRONIZE
}

func NewSynchronizeDataPDU(targetUser uint16) *SynchronizeDataPDU {
	return &SynchronizeDataPDU{
		MessageType: 1,
		TargetUser:  targetUser,
	}
}

type ControlDataPDU struct {
	Action    uint16 `struc:"little"`
	GrantId   uint16 `struc:"little"`
	ControlId uint32 `struc:"little"`
}

func (*ControlDataPDU) Type2() uint8 {
	return PDUTYPE2_CONTROL
}

type FontListDataPDU struct {
	NumberFonts   uint16 `struc:"little"`
	TotalNumFonts uint16 `struc:"little"`
	ListFlags     uint16 `struc:"little"`
	EntrySize     uint16 `struc:"little"`
}

func (*FontListDataPDU) Type2() uint8 {
	return PDUTYPE2_FONTLIST
}

type ErrorInfoDataPDU struct {
	ErrorInfo uint32 `struc:"little"`
}

func (*ErrorInfoDataPDU) Type2() uint8 {
	return PDUTYPE2_SET_ERROR_INFO_PDU
}

type FontMapDataPDU struct {
	NumberEntries   uint16 `struc:"little"`
	TotalNumEntries uint16 `struc:"little"`
	MapFlags        uint16 `struc:"little"`
	EntrySize       uint16 `struc:"little"`
}

func (*FontMapDataPDU) Type2() uint8 {
	return PDUTYPE2_FONTMAP
}

type UpdateData interface {
	FastPathUpdateType() uint8
}

type BitmapCompressedDataHeader struct {
	CbCompFirstRowSize uint16 `struc:"little"`
	CbCompMainBodySize uint16 `struc:"little"`
	CbScanWidth        uint16 `struc:"little"`
	CbUncompressedSize uint16 `struc:"little"`
}

type BitmapData struct {
	DestLeft         uint16 `struc:"little"`
	DestTop          uint16 `struc:"little"`
	DestRight        uint16 `struc:"little"`
	DestBottom       uint16 `struc:"little"`
	Width            uint16 `struc:"little"`
	Height           uint16 `struc:"little"`
	BitsPerPixel     uint16 `struc:"little"`
	Flags            uint16 `struc:"little"`
	BitmapLength     uint16 `struc:"little,sizeof=BitmapDataStream"`
	BitmapComprHdr   *BitmapCompressedDataHeader
	BitmapDataStream []byte
}

type FastPathBitmapUpdateDataPDU struct {
	Header           uint16 `struc:"little"`
	NumberRectangles uint16 `struc:"little,sizeof=Rectangles"`
	Rectangles       []BitmapData
}

func (*FastPathBitmapUpdateDataPDU) FastPathUpdateType() uint8 {
	return FASTPATH_UPDATETYPE_BITMAP
}

type FastPathUpdatePDU struct {
	UpdateHeader     uint8
	CompressionFlags uint8
	Size             uint16
	Data             UpdateData
}

func readFastPathUpdatePDU(r io.Reader) (*FastPathUpdatePDU, error) {
	f := &FastPathUpdatePDU{}
	var err error
	f.UpdateHeader, err = core.ReadUInt8(r)
	if err != nil {
		return nil, err
	}
	f.CompressionFlags, err = core.ReadUInt8(r)
	f.Size, err = core.ReadUint16LE(r)
	if err != nil {
		return nil, err
	}
	dataBytes, err := core.ReadBytes(int(f.Size), r)
	if err != nil {
		return nil, err
	}
	var d UpdateData
	switch f.UpdateHeader & 0xf {
	case FASTPATH_UPDATETYPE_BITMAP:
		d = &FastPathBitmapUpdateDataPDU{}
	default:
		glog.Debug("unsupported FastPathUpdatePDU data type", f.UpdateHeader)
		d = nil
	}
	if d != nil {
		err = struc.Unpack(bytes.NewReader(dataBytes), d)
		if err != nil {
			return nil, err
		}
	}
	f.Data = d
	return f, nil
}

type ShareControlHeader struct {
	TotalLength uint16 `struc:"little"`
	PDUType     uint16 `struc:"little"`
	PDUSource   uint16 `struc:"little"`
}

type PDU struct {
	ShareCtrlHeader *ShareControlHeader
	Message         PDUMessage
}

func NewPDU(userId uint16, message PDUMessage) *PDU {
	pdu := &PDU{}
	pdu.ShareCtrlHeader = &ShareControlHeader{
		TotalLength: uint16(len(message.Serialize()) + 6),
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

	var d PDUMessage
	switch pdu.ShareCtrlHeader.PDUType {
	case PDUTYPE_DEMANDACTIVEPDU:
		d, err = readDemandActivePDU(r)
	case PDUTYPE_DATAPDU:
		d, err = readDataPDU(r)
	case PDUTYPE_CONFIRMACTIVEPDU:
		d, err = readConfirmActivePDU(r)
	case PDUTYPE_DEACTIVATEALLPDU:
		d, err = readDeactiveAllPDU(r)
	default:
		glog.Error("PDU invalid pdu type")
	}
	if err != nil {
		return nil, err
	}
	pdu.Message = d
	return pdu, err
}

func (p *PDU) serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, p.ShareCtrlHeader)
	core.WriteBytes(p.Message.Serialize(), buff)
	return buff.Bytes()
}
