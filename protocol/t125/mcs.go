package t125

import (
	"encoding/asn1"
	"errors"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/core"
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
	MaxChannelIds   int `asn1:"tag:2"`
	MaxUserIds      int `asn1:"tag:2"`
	MaxTokenIds     int `asn1:"tag:2"`
	NumPriorities   int `asn1:"tag:2"`
	MinThoughput    int `asn1:"tag:2"`
	MaxHeight       int `asn1:"tag:2"`
	MaxMCSPDUsize   int `asn1:"tag:2"`
	ProtocolVersion int `asn1:"tag:2"`
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

/**
 * @see http://www.itu.int/rec/T-REC-T.125-199802-I/en page 25
 * @param userData {Buffer}
 * @returns {asn1.univ.Sequence}
 */
type ConnectInitial struct {
	CallingDomainSelector []byte `asn1:"tag:4"`
	CalledDomainSelector  []byte `asn1:"tag:4"`
	UpwardFlag            bool
	TargetParameters      DomainParameters
	MinimumParameters     DomainParameters
	MaximumParameters     DomainParameters
	UserData              []byte `asn1:"application, tag:101"`
}

func NewConnectInitial(userData []byte) *ConnectInitial {
	return &ConnectInitial{[]byte{0x1},
		[]byte{0x1},
		false,
		*NewDomainParameters(34, 2, 0, 1, 0, 1, 0xffff, 2),
		*NewDomainParameters(1, 1, 1, 1, 0, 1, 0x420, 2),
		*NewDomainParameters(0xffff, 0xfc17, 0xffff, 1, 0, 1, 0xffff, 2),
		userData}
	/*userData : new asn1.univ.OctetString(userData)
	}).implicitTag(new asn1.spec.Asn1Tag(asn1.spec.TagClass.Application, asn1.spec.TagFormat.Constructed, 101));*/
}

/**
 * @see http://www.itu.int/rec/T-REC-T.125-199802-I/en page 25
 * @returns {asn1.univ.Sequence}
 */

type ConnectResponse struct {
	result           int `asn1:"tag:10"`
	calledConnectId  int
	domainParameters DomainParameters
	userData         []byte `asn1:"tag:10"`
	//.implicitTag(new asn1.spec.Asn1Tag(asn1.spec.TagClass.Application, asn1.spec.TagFormat.Constructed, 102));
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
	fmt.Println("mcs client on connect", selectedProtocol)
	c.clientCoreData.ServerSelectedProtocol = uint32(selectedProtocol)

	// sendConnectInitial
	conferenceCreateRequest := []byte{}
	connectInitial := NewConnectInitial(conferenceCreateRequest)
	connectInitialBerEncoded, err := asn1.Marshal(connectInitial)
	if err != nil {
		c.Emit("error", errors.New(fmt.Sprintf("mcs sendConnectInitial ber encode error %v", err)))
		return
	}

	_, err = c.transport.Write(connectInitialBerEncoded)
	if err != nil {
		c.Emit("error", errors.New(fmt.Sprintf("mcs sendConnectInitial write error %v", err)))
		return
	}
	c.Once("data", c.recvConnectResponse)
}

func (m *MCSClient) recvConnectResponse(s []byte) {
	fmt.Println("mcs recvConnectResponse", s)
}
