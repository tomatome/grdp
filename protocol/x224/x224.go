package x224

import (
	"bytes"
	"errors"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/protocol"
)

// take idea from https://github.com/Madnikulin50/gordp

/**
 * Message type present in X224 packet header
 */
type MessageType byte

const (
	TPDU_CONNECTION_REQUEST MessageType = 0xE0
	TPDU_CONNECTION_CONFIRM             = 0xD0
	TPDU_DISCONNECT_REQUEST             = 0x80
	TPDU_DATA                           = 0xF0
	TPDU_ERROR                          = 0x70
)

/**
 * Type of negotiation present in negotiation packet
 */
type NegotiationType byte

const (
	TYPE_RDP_NEG_REQ     NegotiationType = 0x01
	TYPE_RDP_NEG_RSP                     = 0x02
	TYPE_RDP_NEG_FAILURE                 = 0x03
)

/**
 * Protocols available for x224 layer
 */
type Protocol uint32

const (
	PROTOCOL_RDP       Protocol = 0x00000000
	PROTOCOL_SSL                = 0x00000001
	PROTOCOL_HYBRID             = 0x00000002
	PROTOCOL_HYBRID_EX          = 0x00000008
)

/**
 * Use to negotiate security layer of RDP stack
 * In node-rdpjs only ssl is available
 * @param opt {object} component type options
 * @see request -> http://msdn.microsoft.com/en-us/library/cc240500.aspx
 * @see response -> http://msdn.microsoft.com/en-us/library/cc240506.aspx
 * @see failure ->http://msdn.microsoft.com/en-us/library/cc240507.aspx
 */

type Negotiation struct {
	Type   NegotiationType
	Flag   uint8
	Length uint16
	Result uint32
}

func NewNegotiation() *Negotiation {
	return &Negotiation{0, 0, 0x0008 /*constant*/, uint32(PROTOCOL_RDP)}
}

//func (x *Negotiation) Write(w core.Writer) {
//	core.WriteByte(byte(x.Type), w)
//	core.WriteUInt8(x.Flag, w)
//	core.WriteUInt16LE(x.Length, w)
//	core.WriteUInt32LE(x.Result, w)
//}
//
//func (x *Negotiation) Read(r core.Reader) error {
//	var err error
//	b, err := core.ReadByte(r)
//	x.Type = NegotiationType(b)
//	x.Flag, err = core.ReadUInt8(r)
//	x.Length, err = core.ReadUInt16LE(r)
//	x.Result, err = core.ReadUInt32LE(r)
//	if x.Length == 0x0008 {
//		return errors.New("invalid x224 negoitiate")
//	}
//	return err
//}

/**
 * X224 client connection request
 * @param opt {object} component type options
 * @see	http://msdn.microsoft.com/en-us/library/cc240470.aspx
 */
type ClientConnectionRequestPDU struct {
	Len         uint8
	Code        MessageType
	Padding1    uint16
	Padding2    uint16
	Padding3    uint8
	Cookie      []byte
	ProtocolNeg Negotiation
	//CorrelationInfo [36]byte
}

func NewClientConnectionRequestPDU(coockie []byte) *ClientConnectionRequestPDU {
	x := ClientConnectionRequestPDU{0, TPDU_CONNECTION_REQUEST, 0, 0, 0,
		coockie, *NewNegotiation() /*, [36]byte{}*/}
	// todo x.Len = uint8(binary.CalcDataLength(&x))
	return &x
}

func (x *ClientConnectionRequestPDU) Serialize() []byte {
	// todo
	var b bytes.Buffer

	return b.Bytes()
}

//func (x *ClientConnectionRequestPDU) Write(w core.Writer) error {
//	core.WriteUInt8(x.Len, w)
//	core.WriteUInt8(uint8(x.Code), w)
//	core.WriteUInt16LE(x.Padding1, w)
//	core.WriteUInt16LE(x.Padding2, w)
//	core.WriteUInt8(x.Padding3, w)
//	w.Write(x.Cookie)
//	core.WriteUInt16LE(0x0a0d, w)
//	x.ProtocolNeg.Write(w)
//	return nil
//}

/**
 * X224 Server connection confirm
 * @param opt {object} component type options
 * @see	http://msdn.microsoft.com/en-us/library/cc240506.aspx
 */
type ServerConnectionConfirm struct {
	Len         uint8
	Code        MessageType
	Padding1    uint16
	Padding2    uint16
	Padding3    uint8
	ProtocolNeg Negotiation
}

/**
 * Header of each data message from x224 layer
 * @returns {type.Component}
 */
type DataHeader struct {
	Header      uint8
	MessageType MessageType
	Separator   uint8
}

func NewDataHeader() *DataHeader {
	return &DataHeader{2, TPDU_DATA /* constant */, 0x80 /*constant*/}
}

/**
 * Common X224 Automata
 * @param presentation {Layer} presentation layer
 */
type X224 struct {
	emission.Emitter
	transport         protocol.Transport
	requestedProtocol Protocol
	selectedProtocol  Protocol
}

func New(t protocol.Transport) *X224 {
	x := &X224{
		*emission.NewEmitter(),
		t,
		PROTOCOL_SSL,
		PROTOCOL_SSL,
	}

	t.On("close", func() {
		x.Emit("close")
	}).On("error", func() {
		x.Emit("error")
	})

	return x
}

func (x *X224) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (x *X224) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (x *X224) Close() error {
	return x.transport.Close()
}

func (x *X224) Connect() error {
	if x.transport == nil {
		return errors.New("no transport")
	}
	message := NewClientConnectionRequestPDU(make([]byte, 0))
	message.ProtocolNeg.Type = TYPE_RDP_NEG_REQ
	message.ProtocolNeg.Result = uint32(x.requestedProtocol)

	_, err := x.transport.Write(message.Serialize())

	x.transport.Once("data", func() {
		x.recvConnectionConfirm()
	})

	return err
}

func (x *X224) recvConnectionConfirm() {

}
