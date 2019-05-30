package t125

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/t125/ber"
	"github.com/icodeface/grdp/protocol/t125/gcc"
	"github.com/icodeface/grdp/protocol/x224"
)

// take idea from https://github.com/Madnikulin50/gordp

// Multiple Channel Service layer

type MCSMessage uint8

const (
	MCS_TYPE_CONNECT_INITIAL  MCSMessage = 0x65
	MCS_TYPE_CONNECT_RESPONSE            = 0x66
)

type MCSDomainPDU uint16

const (
	ERECT_DOMAIN_REQUEST          MCSDomainPDU = 1
	DISCONNECT_PROVIDER_ULTIMATUM              = 8
	ATTACH_USER_REQUEST                        = 10
	ATTACH_USER_CONFIRM                        = 11
	CHANNEL_JOIN_REQUEST                       = 14
	CHANNEL_JOIN_CONFIRM                       = 15
	SEND_DATA_REQUEST                          = 25
	SEND_DATA_INDICATION                       = 26
)

type MCSChannel uint16

const (
	MCS_GLOBAL_CHANNEL   MCSChannel = 1003
	MCS_USERCHANNEL_BASE            = 1001
)

type DomainParameters struct {
	MaxChannelIds   int `asn1: "tag:2"`
	MaxUserIds      int `asn1: "tag:2"`
	MaxTokenIds     int `asn1: "tag:2"`
	NumPriorities   int `asn1: "tag:2"`
	MinThoughput    int `asn1: "tag:2"`
	MaxHeight       int `asn1: "tag:2"`
	MaxMCSPDUsize   int `asn1: "tag:2"`
	ProtocolVersion int `asn1: "tag:2"`
}

/**
 * @see http://www.itu.int/rec/T-REC-T.125-199802-I/en page 25
 * @returns {asn1.univ.Sequence}
 */
func NewDomainParameters(maxChannelIds int,
	maxUserIds int,
	maxTokenIds int,
	numPriorities int,
	minThoughput int,
	maxHeight int,
	maxMCSPDUsize int,
	protocolVersion int) *DomainParameters {
	return &DomainParameters{maxChannelIds, maxUserIds, maxTokenIds,
		numPriorities, minThoughput, maxHeight, maxMCSPDUsize, protocolVersion}
}

func (d *DomainParameters) BER() []byte {
	buff := &bytes.Buffer{}
	ber.WriteInteger(d.MaxChannelIds, buff)
	ber.WriteInteger(d.MaxUserIds, buff)
	ber.WriteInteger(d.MaxTokenIds, buff)
	ber.WriteInteger(1, buff)
	ber.WriteInteger(0, buff)
	ber.WriteInteger(1, buff)
	ber.WriteInteger(d.MaxMCSPDUsize, buff)
	ber.WriteInteger(2, buff)
	return buff.Bytes()
}

/**
 * @see http://www.itu.int/rec/T-REC-T.125-199802-I/en page 25
 * @param userData {Buffer}
 * @returns {asn1.univ.Sequence}
 */
type ConnectInitial struct {
	CallingDomainSelector []byte `asn1: "tag:4"`
	CalledDomainSelector  []byte `asn1: "tag:4"`
	UpwardFlag            bool
	TargetParameters      DomainParameters
	MinimumParameters     DomainParameters
	MaximumParameters     DomainParameters
	UserData              []byte `asn1: "application, tag:101"`
}

func NewConnectInitial(userData []byte) ConnectInitial {
	return ConnectInitial{[]byte{0x1},
		[]byte{0x1},
		true,
		*NewDomainParameters(34, 2, 0, 1, 0, 1, 0xffff, 2),
		*NewDomainParameters(1, 1, 1, 1, 0, 1, 0x420, 2),
		*NewDomainParameters(0xffff, 0xfc17, 0xffff, 1, 0, 1, 0xffff, 2),
		userData}
}

func (c *ConnectInitial) BER() []byte {
	buff := &bytes.Buffer{}
	ber.WriteOctetstring(string(c.CallingDomainSelector), buff)
	ber.WriteOctetstring(string(c.CalledDomainSelector), buff)
	ber.WriteBoolean(c.UpwardFlag, buff)
	ber.WriteEncodedDomainParams(c.TargetParameters.BER(), buff)
	ber.WriteEncodedDomainParams(c.MinimumParameters.BER(), buff)
	ber.WriteEncodedDomainParams(c.MaximumParameters.BER(), buff)
	ber.WriteOctetstring(string(c.UserData), buff)
	return buff.Bytes()
}

/**
 * @see http://www.itu.int/rec/T-REC-T.125-199802-I/en page 25
 * @returns {asn1.univ.Sequence}
 */

type ConnectResponse struct {
	result           int `asn1: "tag:10"`
	calledConnectId  int
	domainParameters DomainParameters
	userData         []byte `asn1: "tag:10"`
}

func NewConnectResponse(userData []byte) *ConnectResponse {
	return &ConnectResponse{0,
		0,
		*NewDomainParameters(22, 3, 0, 1, 0, 1, 0xfff8, 2),
		userData}
}

type MCSChannelInfo struct {
	id   MCSChannel
	name string
}

type MCS struct {
	emission.Emitter
	transport  core.Transport
	recvOpCode MCSDomainPDU
	sendOpCode MCSDomainPDU
	channels   []MCSChannelInfo
}

func NewMCS(t core.Transport, recvOpCode MCSDomainPDU, sendOpCode MCSDomainPDU) *MCS {
	m := &MCS{
		*emission.NewEmitter(),
		t,
		recvOpCode,
		sendOpCode,
		[]MCSChannelInfo{{MCS_GLOBAL_CHANNEL, "global"}},
	}

	m.transport.On("close", func() {
		m.Emit("close")
	}).On("error", func(err error) {
		m.Emit("error", err)
	})
	return m
}

func (x *MCS) Read(b []byte) (n int, err error) {
	return x.transport.Read(b)
}

func (x *MCS) Write(b []byte) (n int, err error) {
	return x.transport.Write(b)
}

func (m *MCS) Close() error {
	return m.transport.Close()
}

type MCSClient struct {
	*MCS
	clientCoreData     *gcc.ClientCoreData
	clientNetworkData  *gcc.ClientNetworkData
	clientSecurityData *gcc.ClientSecurityData

	serverCoreData     *gcc.ServerCoreData
	serverNetworkData  *gcc.ServerNetworkData
	serverSecurityData *gcc.ServerSecurityData

	channelsConnected int
	userId            uint16
}

func NewMCSClient(t core.Transport) *MCSClient {
	c := &MCSClient{
		MCS:                NewMCS(t, SEND_DATA_INDICATION, SEND_DATA_REQUEST),
		clientCoreData:     gcc.NewClientCoreData(),
		clientNetworkData:  gcc.NewClientNetworkData(),
		clientSecurityData: gcc.NewClientSecurityData(),
	}
	c.transport.On("connect", c.connect)
	return c
}

func (c *MCSClient) connect(selectedProtocol x224.Protocol) {
	glog.Debug("mcs client on connect", selectedProtocol)
	c.clientCoreData.ServerSelectedProtocol = uint32(selectedProtocol)

	// sendConnectInitial
	userDataBuff := bytes.Buffer{}
	userDataBuff.Write(c.clientCoreData.Block())
	userDataBuff.Write(c.clientNetworkData.Block())
	userDataBuff.Write(c.clientSecurityData.Block())

	ccReq := gcc.MakeConferenceCreateRequest(userDataBuff.Bytes())
	connectInitial := NewConnectInitial(ccReq)
	connectInitialBerEncoded := connectInitial.BER()

	dataBuff := &bytes.Buffer{}
	ber.WriteApplicationTag(uint8(MCS_TYPE_CONNECT_INITIAL), len(connectInitialBerEncoded), dataBuff)
	dataBuff.Write(connectInitialBerEncoded)

	_, err := c.transport.Write(dataBuff.Bytes())
	if err != nil {
		c.Emit("error", errors.New(fmt.Sprintf("mcs sendConnectInitial write error %v", err)))
		return
	}
	glog.Debug("mcs wait for data event")
	c.transport.Once("data", c.recvConnectResponse)
}

func (m *MCSClient) recvConnectResponse(s []byte) {
	glog.Debug("mcs recvConnectResponse", s)
	// todo

	// record server gcc block

	// send domain request

	// send attach user request
	m.transport.Once("data", m.recvAttachUserConfirm)

}

func (m *MCSClient) recvAttachUserConfirm(s []byte) {
	glog.Debug("mcs recvAttachUserConfirm")
	// todo

	//ask channel for specific user

	// channel connect automata
	m.connectChannels([]byte{})
}

func (m *MCSClient) connectChannels(s []byte) {
	// todo

	if m.channelsConnected == len(m.channels) {
		m.transport.On("data", func(s []byte) {

		})
		// send client and sever gcc informations
		// callback to sec
		m.Emit("connect", m.userId, m.channels)
	}

	// sendChannelJoinRequest

	m.transport.Once("data", m.recvChannelJoinConfirm)
}

func (m *MCSClient) recvChannelJoinConfirm(s []byte) {
	// todo
	m.connectChannels(s)
}
