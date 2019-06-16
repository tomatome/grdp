package pdu

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/emission"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/t125/gcc"
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
	core.WriteUInt16LE(d.NumberCapabilities, buff)
	core.WriteUInt16LE(d.Pad2Octets, buff)
	for _, cap := range d.CapabilitySets {
		core.WriteBytes(cap.Serialize(), buff)
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
	for i := 0; i < int(d.NumberCapabilities); i++ {
		capType, err := core.ReadUint16LE(r)
		if err != nil {
			return nil, err
		}
		capLen, err := core.ReadUint16LE(r)
		if err != nil {
			return nil, err
		}
		core.ReadBytes(int(capLen)-4, r)
		fmt.Println("cap type is", capType, "len is", capLen)
		switch CapsType(capType) {
		// todo
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
	Pad2Octets                 uint16       `struc:"little"`
	CapabilitySets             []Capability `struc:"sizefrom=NumberCapabilities"`
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

type PDULayer struct {
	emission.Emitter
	transport core.Transport
	sharedId  uint32
	userId    uint16
}

func NewPDULayer(t core.Transport) *PDULayer {
	p := &PDULayer{
		Emitter:   *emission.NewEmitter(),
		transport: t,
		sharedId:  0x103EA,
	}

	t.On("close", func() {
		p.Emit("close")
	}).On("error", func(err error) {
		p.Emit("error", err)
	})
	return p
}

func (p *PDULayer) sendPDU(message PDUMessage) {
	pdu := NewPDU(p.userId, message)
	p.transport.Write(pdu.serialize())
}

type Client struct {
	*PDULayer
	clientCoreData     *gcc.ClientCoreData
	serverCapabilities map[CapsType]Capability
}

func NewClient(t core.Transport) *Client {
	c := &Client{
		PDULayer:           NewPDULayer(t),
		serverCapabilities: make(map[CapsType]Capability, 0),
	}
	c.transport.Once("connect", c.connect)
	return c
}

func (c *Client) connect(data *gcc.ClientCoreData, userId uint16) {
	glog.Debug("pdu connect")
	c.clientCoreData = data
	c.userId = userId
	c.transport.Once("data", c.recvDemandActivePDU)
}

func (c *Client) recvDemandActivePDU(s []byte) {
	glog.Debug("PDU recvDemandActivePDU", hex.EncodeToString(s))
	r := bytes.NewReader(s)
	pdu, err := readPDU(r)
	if err != nil {
		glog.Error(err)
		return
	}
	if pdu.ShareCtrlHeader.PDUType != PDUTYPE_DEMANDACTIVEPDU {
		glog.Info("PDU ignore message during connection sequence, type is", pdu.ShareCtrlHeader.PDUType)
		c.transport.Once("data", c.recvDemandActivePDU)
		return
	}
	c.sharedId = pdu.Message.(*DemandActivePDU).SharedId

	fmt.Println("CapabilitySets:", pdu.Message.(*DemandActivePDU).CapabilitySets)
	for _, caps := range pdu.Message.(*DemandActivePDU).CapabilitySets {
		c.serverCapabilities[caps.Type()] = caps
	}

	c.sendConfirmActivePDU()
	c.sendClientFinalizeSynchronizePDU()
	c.transport.Once("data", c.recvServerSynchronizePDU)
}

func (c *Client) sendConfirmActivePDU() {
	glog.Debug("PDU sendConfirmActivePDU")
	// todo
}

func (c *Client) sendClientFinalizeSynchronizePDU() {
	glog.Debug("PDU sendClientFinalizeSynchronizePDU")
	// todo
}

func (c *Client) recvServerSynchronizePDU(s []byte) {
	glog.Debug("PDU recvServerSynchronizePDU")
	// todo
}

func (c *Client) recvServerControlCooperatePDU() {
	// todo
}

func (c *Client) recvPDU() {
	// todo
}
