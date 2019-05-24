package x224

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/protocol/tpkt"
	"io"
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

func (x *Negotiation) Serialize() []byte {
	buff := &bytes.Buffer{}
	core.WriteByte(byte(x.Type), buff) // 1
	core.WriteUInt8(x.Flag, buff)      // 0
	core.WriteUInt16LE(x.Length, buff) // 8 0
	core.WriteUInt32LE(x.Result, buff) // 1 0 0 0 vs 3 0 0 0
	return buff.Bytes()
}

func ReadNegotiation(r io.Reader) (*Negotiation, error) {
	// 3 0 8 0 5 0 0 0
	n := &Negotiation{}

	b, err := core.ReadByte(r) // 3 TYPE_RDP_NEG_FAILURE
	if err != nil {
		return nil, err
	}
	n.Type = NegotiationType(b)

	n.Flag, err = core.ReadUInt8(r)      // 0
	n.Length, err = core.ReadUint16LE(r) // 8
	n.Result, err = core.ReadUInt32LE(r) // 0 5

	return n, nil
}

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
	x.Len = uint8(len(x.Serialize()) - 1)
	return &x
}

func (x *ClientConnectionRequestPDU) Serialize() []byte {
	buff := &bytes.Buffer{}
	core.WriteUInt8(x.Len, buff)
	core.WriteUInt8(uint8(x.Code), buff)
	core.WriteUInt16LE(x.Padding1, buff)
	core.WriteUInt16LE(x.Padding2, buff)
	core.WriteUInt8(x.Padding3, buff)
	buff.Write(x.Cookie)
	//if x.Len > 14:
	core.WriteUInt16LE(0x0A0D, buff)
	buff.Write(x.ProtocolNeg.Serialize())
	fmt.Println("ProtocolNeg", x.ProtocolNeg.Serialize())
	return buff.Bytes()
}

// 14 224 0 0 0 0 0
// 1 0
// 8 0 3 0 0 0

// 16 224 0 0 0 0 0
//
// 10 13
//
// 1 0
// 8 0 1 0 0 0

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

func ReadServerConnectionConfirm(r io.Reader) (*ServerConnectionConfirm, error) {
	// 14 208 0 0 18 52 0 3 0 8 0 5 0 0 0
	s := &ServerConnectionConfirm{}
	var err error
	s.Len, err = core.ReadUInt8(r) // 14

	code, err := core.ReadUInt8(r) // 208 TPDU_CONNECTION_CONFIRM
	s.Code = MessageType(code)

	s.Padding1, err = core.ReadUint16LE(r) // 0 0
	s.Padding2, err = core.ReadUint16LE(r) // 18 52
	s.Padding3, err = core.ReadUInt8(r)    // 0

	neo, err := ReadNegotiation(r)
	if err != nil {
		return nil, err
	}
	s.ProtocolNeg = *neo

	return s, err
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
	transport         core.Transport
	requestedProtocol Protocol
	selectedProtocol  Protocol
}

func New(t core.Transport) *X224 {
	x := &X224{
		*emission.NewEmitter(),
		t,
		PROTOCOL_SSL,
		PROTOCOL_SSL,
	}

	t.On("close", func() {
		x.Emit("close")
	}).On("error", func(err error) {
		x.Emit("error", err)
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
	fmt.Println("x224 Connect")
	if x.transport == nil {
		return errors.New("no transport")
	}
	message := NewClientConnectionRequestPDU(make([]byte, 0))
	message.ProtocolNeg.Type = TYPE_RDP_NEG_REQ
	message.ProtocolNeg.Result = uint32(x.requestedProtocol)

	_, err := x.transport.Write(message.Serialize())
	x.transport.Once("data", x.recvConnectionConfirm)
	return err
}

func (x *X224) recvConnectionConfirm(s []byte) {
	fmt.Println("x224 recvConnectionConfirm", hex.EncodeToString(s))

	// rdpy: 0ed000001234000209080002000000
	// we:   0ed000001234000300080005000000
	// we2:  0ed000001234000209080001000000

	message, err := ReadServerConnectionConfirm(bytes.NewReader(s))
	if err != nil {
		fmt.Println("ReadServerConnectionConfirm err", err)
		return
	}

	if message.ProtocolNeg.Type == TYPE_RDP_NEG_FAILURE {
		fmt.Println("NODE_RDP_PROTOCOL_NEG_FAILURE")
		return
	}

	if message.ProtocolNeg.Type == TYPE_RDP_NEG_RSP {
		fmt.Println("TYPE_RDP_NEG_RSP")
		x.selectedProtocol = Protocol(message.ProtocolNeg.Result)
	}

	if x.selectedProtocol == PROTOCOL_HYBRID || x.selectedProtocol == PROTOCOL_HYBRID_EX {
		fmt.Println("NODE_RDP_PROTOCOL_NLA_NOT_SUPPORTED")
		return
	}

	if x.selectedProtocol == PROTOCOL_RDP {
		fmt.Println("RDP standard security selected")
		return
	}

	x.transport.On("data", x.recvData)

	if x.selectedProtocol == PROTOCOL_SSL {
		fmt.Println("SSL standard security selected")
		err := x.transport.(*tpkt.TPKT).Conn.StartTLS()
		if err != nil {
			fmt.Println("start tls failed", err)
		}
	}
}

func (x *X224) recvData(s []byte) {
	fmt.Println("x224 recvData", s)
	// check header
	//x224DataHeader().read(s);
	//this.emit('data', s);
}
