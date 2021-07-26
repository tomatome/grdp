package nla

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"time"

	"github.com/lunixbochs/struc"
	"github.com/tomatome/grdp/core"
	"github.com/tomatome/grdp/glog"
)

const (
	MsvAvEOL             = 0x0000
	MsvAvNbComputerName  = 0x0001
	MsvAvNbDomainName    = 0x0002
	MsvAvDnsComputerName = 0x0003
	MsvAvDnsDomainName   = 0x0004
	MsvAvDnsTreeName     = 0x0005
	MsvAvFlags           = 0x0006
	MsvAvTimestamp       = 0x0007
	MsvAvSingleHost      = 0x0008
	MsvAvTargetName      = 0x0009
	MsvChannelBindings   = 0x000A
)

type AVPair struct {
	Id    uint16 `struc:"little"`
	Len   uint16 `struc:"little,sizeof=Value"`
	Value []byte `struc:"little"`
}

const (
	NTLMSSP_NEGOTIATE_56                       = 0x80000000
	NTLMSSP_NEGOTIATE_KEY_EXCH                 = 0x40000000
	NTLMSSP_NEGOTIATE_128                      = 0x20000000
	NTLMSSP_NEGOTIATE_VERSION                  = 0x02000000
	NTLMSSP_NEGOTIATE_TARGET_INFO              = 0x00800000
	NTLMSSP_REQUEST_NON_NT_SESSION_KEY         = 0x00400000
	NTLMSSP_NEGOTIATE_IDENTIFY                 = 0x00100000
	NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY = 0x00080000
	NTLMSSP_TARGET_TYPE_SERVER                 = 0x00020000
	NTLMSSP_TARGET_TYPE_DOMAIN                 = 0x00010000
	NTLMSSP_NEGOTIATE_ALWAYS_SIGN              = 0x00008000
	NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED = 0x00002000
	NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED      = 0x00001000
	NTLMSSP_NEGOTIATE_NTLM                     = 0x00000200
	NTLMSSP_NEGOTIATE_LM_KEY                   = 0x00000080
	NTLMSSP_NEGOTIATE_DATAGRAM                 = 0x00000040
	NTLMSSP_NEGOTIATE_SEAL                     = 0x00000020
	NTLMSSP_NEGOTIATE_SIGN                     = 0x00000010
	NTLMSSP_REQUEST_TARGET                     = 0x00000004
	NTLM_NEGOTIATE_OEM                         = 0x00000002
	NTLMSSP_NEGOTIATE_UNICODE                  = 0x00000001
)

type NVersion struct {
	ProductMajorVersion uint8   `struc:"little"`
	ProductMinorVersion uint8   `struc:"little"`
	ProductBuild        uint16  `struc:"little"`
	Reserved            [3]byte `struc:"little"`
	UInt8               uint8   `struc:"little"`
}

type Message interface {
	Serialize() []byte
}

type NegotiateMessage struct {
	Signature               [8]byte  `struc:"little"`
	MessageType             uint32   `struc:"little"`
	NegotiateFlags          uint32   `struc:"little"`
	DomainNameLen           uint16   `struc:"little"`
	DomainNameMaxLen        uint16   `struc:"little"`
	DomainNameBufferOffset  uint32   `struc:"little"`
	WorkstationLen          uint16   `struc:"little"`
	WorkstationMaxLen       uint16   `struc:"little"`
	WorkstationBufferOffset uint32   `struc:"little"`
	Version                 NVersion `struc:"little"`
	Payload                 [32]byte `struc:"skip"`
}

func NewNegotiateMessage() *NegotiateMessage {
	return &NegotiateMessage{
		Signature:   [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0x00},
		MessageType: 0x00000001,
	}
}

func (m *NegotiateMessage) Serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, m)
	res := buff.Bytes()
	if (m.NegotiateFlags & NTLMSSP_NEGOTIATE_VERSION) <= 0 {
		res = append(res[0:32], res[40:]...)
	}
	return res
}

type ChallengeMessage struct {
	Signature              [8]byte  `struc:"little"`
	MessageType            uint32   `struc:"little"`
	TargetNameLen          uint16   `struc:"little"`
	TargetNameMaxLen       uint16   `struc:"little"`
	TargetNameBufferOffset uint32   `struc:"little"`
	NegotiateFlags         uint32   `struc:"little"`
	ServerChallenge        [8]byte  `struc:"little"`
	Reserved               [8]byte  `struc:"little"`
	TargetInfoLen          uint16   `struc:"little"`
	TargetInfoMaxLen       uint16   `struc:"little"`
	TargetInfoBufferOffset uint32   `struc:"little"`
	Version                NVersion `struc:"skip"`
	Payload                []byte   `struc:"skip"`
}

// total len - payload len
func (m *ChallengeMessage) BaseLen() uint32 {
	return 56
}

func (m *ChallengeMessage) getTargetInfo() []byte {
	if m.TargetInfoLen == 0 {
		return make([]byte, 0)
	}
	offset := m.BaseLen()
	start := m.TargetInfoBufferOffset - offset
	return m.Payload[start : start+uint32(m.TargetInfoLen)]
}

func (m *ChallengeMessage) getTargetInfoTimestamp(data []byte) []byte {
	r := bytes.NewReader(data)
	for r.Len() > 0 {
		avPair := &AVPair{}
		struc.Unpack(r, avPair)

		if avPair.Id == MsvAvTimestamp {
			return avPair.Value
		}

		if avPair.Id == MsvAvEOL {
			break
		}
	}
	return nil
}

func (m *ChallengeMessage) Serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, m)
	buff.Write(m.Payload)
	return buff.Bytes()
}

func NewChallengeMessage() *ChallengeMessage {
	return &ChallengeMessage{
		Signature:   [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0x00},
		MessageType: 0x00000002,
	}
}

type AuthenticateMessage struct {
	Signature                          [8]byte
	MessageType                        uint32 `struc:"little"`
	LmChallengeResponseLen             uint16 `struc:"little"`
	LmChallengeResponseMaxLen          uint16 `struc:"little"`
	LmChallengeResponseBufferOffset    uint32 `struc:"little"`
	NtChallengeResponseLen             uint16 `struc:"little"`
	NtChallengeResponseMaxLen          uint16 `struc:"little"`
	NtChallengeResponseBufferOffset    uint32 `struc:"little"`
	DomainNameLen                      uint16 `struc:"little"`
	DomainNameMaxLen                   uint16 `struc:"little"`
	DomainNameBufferOffset             uint32 `struc:"little"`
	UserNameLen                        uint16 `struc:"little"`
	UserNameMaxLen                     uint16 `struc:"little"`
	UserNameBufferOffset               uint32 `struc:"little"`
	WorkstationLen                     uint16 `struc:"little"`
	WorkstationMaxLen                  uint16 `struc:"little"`
	WorkstationBufferOffset            uint32 `struc:"little"`
	EncryptedRandomSessionLen          uint16 `struc:"little"`
	EncryptedRandomSessionMaxLen       uint16 `struc:"little"`
	EncryptedRandomSessionBufferOffset uint32 `struc:"little"`
	NegotiateFlags                     uint32 `struc:"little"`
	Version                            NVersion
	MIC                                [16]byte
	Payload                            []byte `struc:"skip"`
}

func (m *AuthenticateMessage) BaseLen() uint32 {
	return 88
}

func NewAuthenticateMessage(negFlag uint32, domain, user, workstation string,
	lmchallResp, ntchallResp, enRandomSessKey []byte) *AuthenticateMessage {
	msg := &AuthenticateMessage{
		Signature:      [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0x00},
		MessageType:    0x00000003,
		NegotiateFlags: negFlag,
	}
	payloadBuff := &bytes.Buffer{}

	domainBytes := []byte(domain)
	userBytes := []byte(user)
	wsBytes := []byte(workstation)
	if negFlag&NTLMSSP_NEGOTIATE_UNICODE != 0 {
		domainBytes = UnicodeEncode(domain)
		userBytes = UnicodeEncode(user)
		wsBytes = UnicodeEncode(workstation)
	}
	msg.DomainNameLen = uint16(len(domainBytes))
	msg.DomainNameBufferOffset = msg.BaseLen()
	payloadBuff.Write(domainBytes)

	msg.UserNameLen = uint16(len(userBytes))
	msg.UserNameBufferOffset = msg.DomainNameBufferOffset + uint32(msg.DomainNameLen)
	payloadBuff.Write(userBytes)

	msg.WorkstationLen = uint16(len(wsBytes))
	msg.WorkstationBufferOffset = msg.UserNameBufferOffset + uint32(msg.UserNameLen)
	payloadBuff.Write(wsBytes)

	msg.LmChallengeResponseLen = uint16(len(lmchallResp))
	msg.LmChallengeResponseBufferOffset = msg.WorkstationBufferOffset + uint32(msg.WorkstationLen)
	payloadBuff.Write(lmchallResp)

	msg.NtChallengeResponseLen = uint16(len(ntchallResp))
	msg.NtChallengeResponseBufferOffset = msg.LmChallengeResponseBufferOffset + uint32(msg.LmChallengeResponseLen)
	payloadBuff.Write(ntchallResp)

	msg.EncryptedRandomSessionLen = uint16(len(enRandomSessKey))
	msg.EncryptedRandomSessionBufferOffset = msg.NtChallengeResponseBufferOffset + uint32(msg.NtChallengeResponseLen)
	payloadBuff.Write(enRandomSessKey)

	msg.Payload = payloadBuff.Bytes()
	return msg
}

func (m *AuthenticateMessage) Serialize() []byte {
	buff := &bytes.Buffer{}
	struc.Pack(buff, m)
	buff.Write(m.Payload)
	res := buff.Bytes()
	return res
}

type NTLMv2 struct {
	domain              string
	user                string
	password            string
	respKeyNT           []byte
	respKeyLM           []byte
	negotiateMessage    *NegotiateMessage
	challengeMessage    *ChallengeMessage
	authenticateMessage *AuthenticateMessage
}

func NewNTLMv2(domain, user, password string) *NTLMv2 {
	return &NTLMv2{
		domain:    domain,
		user:      user,
		password:  password,
		respKeyNT: NTOWFv2(password, user, domain),
		respKeyLM: LMOWFv2(password, user, domain),
	}
}

// generate first handshake messgae
func (n *NTLMv2) GetNegotiateMessage() *NegotiateMessage {
	negoMsg := NewNegotiateMessage()
	negoMsg.NegotiateFlags = NTLMSSP_NEGOTIATE_KEY_EXCH |
		NTLMSSP_NEGOTIATE_128 |
		NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
		NTLMSSP_NEGOTIATE_NTLM |
		NTLMSSP_NEGOTIATE_SEAL |
		NTLMSSP_NEGOTIATE_SIGN |
		NTLMSSP_REQUEST_TARGET |
		NTLMSSP_NEGOTIATE_UNICODE
	n.negotiateMessage = negoMsg
	return n.negotiateMessage
}

//  process NTLMv2 Authenticate hash
func (n *NTLMv2) ComputeResponse(respKeyNT, respKeyLM, serverChallenge, clientChallenge,
	timestamp, serverName []byte) (ntChallResp, lmChallResp, SessBaseKey []byte) {

	tempBuff := &bytes.Buffer{}
	tempBuff.Write([]byte{0x01, 0x01}) // Responser version, HiResponser version
	tempBuff.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	tempBuff.Write(timestamp)
	tempBuff.Write(clientChallenge)
	tempBuff.Write([]byte{0x00, 0x00, 0x00, 0x00})
	tempBuff.Write(serverName)

	ntBuf := bytes.NewBuffer(serverChallenge)
	ntBuf.Write(tempBuff.Bytes())
	ntProof := HMAC_MD5(respKeyNT, ntBuf.Bytes())

	ntChallResp = make([]byte, 0, len(ntProof)+tempBuff.Len())
	ntChallResp = append(ntChallResp, ntProof...)
	ntChallResp = append(ntChallResp, tempBuff.Bytes()...)

	lmBuf := bytes.NewBuffer(serverChallenge)
	lmBuf.Write(clientChallenge)
	lmChallResp = HMAC_MD5(respKeyLM, lmBuf.Bytes())
	lmChallResp = append(lmChallResp, clientChallenge...)

	SessBaseKey = HMAC_MD5(respKeyNT, ntProof)
	return
}

func MIC(exportedSessionKey []byte, negotiateMessage, challengeMessage, authenticateMessage Message) []byte {
	buff := bytes.Buffer{}
	buff.Write(negotiateMessage.Serialize())
	buff.Write(challengeMessage.Serialize())
	buff.Write(authenticateMessage.Serialize())
	return HMAC_MD5(exportedSessionKey, buff.Bytes())
}

func SIGNKEY(exportedSessionKey []byte, isClient bool) []byte {
	buff := bytes.NewBuffer(exportedSessionKey)
	if isClient {
		buff.WriteString("session key to client-to-server signing key magic constant\x00")
	} else {
		buff.WriteString("session key to server-to-client signing key magic constant\x00")
	}
	return MD5(buff.Bytes())
}

func (n *NTLMv2) GetAuthenticateMessage(s []byte) (*AuthenticateMessage, *NTLMv2Security) {
	challengeMsg := &ChallengeMessage{}
	r := bytes.NewReader(s)
	err := struc.Unpack(r, challengeMsg)
	if err != nil {
		glog.Error("read challengeMsg", err)
		return nil, nil
	}
	if challengeMsg.NegotiateFlags&NTLMSSP_NEGOTIATE_VERSION != 0 {
		version := NVersion{}
		err := struc.Unpack(r, &version)
		if err != nil {
			glog.Error("read version", err)
			return nil, nil
		}
		challengeMsg.Version = version
	}
	challengeMsg.Payload, _ = core.ReadBytes(r.Len(), r)
	n.challengeMessage = challengeMsg

	serverName := challengeMsg.getTargetInfo()
	timestamp := challengeMsg.getTargetInfoTimestamp(serverName)
	computeMIC := false
	if timestamp == nil {
		b := &bytes.Buffer{}
		struc.Pack(b, time.Now().UnixNano())
		timestamp = b.Bytes()
	} else {
		computeMIC = true
	}

	serverChallenge := challengeMsg.ServerChallenge[:]
	clientChallenge := core.Random(8)

	ntChallengeResponse, lmChallengeResponse, SessionBaseKey := n.ComputeResponse(
		n.respKeyNT, n.respKeyLM, serverChallenge, clientChallenge, timestamp, serverName)

	exchangeKey := SessionBaseKey
	exportedSessionKey := core.Random(16)
	EncryptedRandomSessionKey := make([]byte, len(exportedSessionKey))
	rc, _ := rc4.NewCipher(exchangeKey)
	rc.XORKeyStream(EncryptedRandomSessionKey, exportedSessionKey)

	n.authenticateMessage = NewAuthenticateMessage(challengeMsg.NegotiateFlags,
		n.domain, n.user, "", lmChallengeResponse, ntChallengeResponse, EncryptedRandomSessionKey)

	if computeMIC {
		copy(n.authenticateMessage.MIC[:], MIC(exportedSessionKey, n.negotiateMessage, n.challengeMessage, n.authenticateMessage)[:16])
	}

	md := md5.New()
	//ClientSigningKey
	md.Write(exportedSessionKey)
	ClientSigningKey := md.Sum([]byte("client-to-server signing"))
	//ServerSigningKey
	md.Reset()
	md.Write(exportedSessionKey)
	ServerSigningKey := md.Sum([]byte("server-to-client signing"))
	//ClientSealingKey
	md.Reset()
	md.Write(exportedSessionKey)
	ClientSealingKey := md.Sum([]byte("client-to-server sealing"))
	//ServerSealingKey
	md.Reset()
	md.Write(exportedSessionKey)
	ServerSealingKey := md.Sum([]byte("server-to-client sealing"))

	encryptRC4, _ := rc4.NewCipher(ClientSealingKey)
	decryptRC4, _ := rc4.NewCipher(ServerSealingKey)

	ntlmSec := &NTLMv2Security{encryptRC4, decryptRC4, ClientSigningKey, ServerSigningKey, 0}

	return n.authenticateMessage, ntlmSec
}

func (n *NTLMv2) GetEncodedCredentials() (string, string, string) {
	return n.domain, n.user, n.password
}

type NTLMv2Security struct {
	EncryptRC4 *rc4.Cipher
	DecryptRC4 *rc4.Cipher
	SigningKey []byte
	VerifyKey  []byte
	SeqNum     uint32
}

func (n *NTLMv2Security) GssEncrypt(s []byte) []byte {
	p := make([]byte, len(s))
	n.EncryptRC4.Reset()
	n.EncryptRC4.XORKeyStream(p, s)

	b := &bytes.Buffer{}

	//signature
	core.WriteUInt32LE(n.SeqNum, b)
	core.WriteBytes(s, b)
	s1 := HMAC_MD5(n.SigningKey, b.Bytes())[:8]
	checksum := make([]byte, 8)
	n.EncryptRC4.Reset()
	n.EncryptRC4.XORKeyStream(checksum, s1)

	b.Reset()
	core.WriteUInt32LE(0x00000001, b)
	core.WriteBytes(checksum, b)
	core.WriteUInt32LE(n.SeqNum, b)

	core.WriteBytes(p, b)

	n.SeqNum++

	return b.Bytes()
}
func (n *NTLMv2Security) GssDecrypt(s []byte) []byte {
	r := bytes.NewReader(s)
	core.ReadUInt32LE(r) //version
	checksum, _ := core.ReadBytes(8, r)
	seqNum, _ := core.ReadUInt32LE(r)
	data, _ := core.ReadBytes(r.Len(), r)

	p := make([]byte, len(data))
	n.DecryptRC4.Reset()
	n.DecryptRC4.XORKeyStream(p, data)

	n.DecryptRC4.Reset()
	check := make([]byte, len(checksum))
	n.DecryptRC4.XORKeyStream(check, checksum)

	b := &bytes.Buffer{}
	core.WriteUInt32LE(seqNum, b)
	core.WriteBytes(p, b)
	verify := HMAC_MD5(n.VerifyKey, b.Bytes())
	if string(verify) != string(check) {
		return nil
	}
	return p
}
