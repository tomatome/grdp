package nla

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
	ProductMajorVersion uint8
	ProductMinorVersion uint8
	ProductBuild        uint16
	Reserved            [3]byte
	UInt8               uint8
}

type NegotiateMessage struct {
	Signature               [8]byte
	MessageType             uint32
	NegotiateFlags          uint32
	DomainNameLen           uint16
	DomainNameMaxLen        uint16
	DomainNameBufferOffset  uint32
	WorkstationLen          uint16
	WorkstationMaxLen       uint16
	WorkstationBufferOffset uint32
	Varsion                 NVersion
	Payload                 string
}

type ChallengeMessage struct {
	Signature              [8]byte
	MessageType            uint32
	TargetNameLen          uint16
	TargetNameMaxLen       uint16
	TargetNameBufferOffset uint32
	NegotiateFlags         uint32
	ServerChallenge        [8]byte
	Reserved               [8]byte
	TargetInfoLen          uint16
	TargetInfoMaxLen       uint16
	TargetInfoBufferOffset uint32
	Version                NVersion
	Payload                string
}

type AuthenticateMessage struct {
	Signature                          [8]byte
	MessageType                        uint32
	LmChallengeResponseLen             uint16
	LmChallengeResponseMaxLen          uint16
	LmChallengeResponseBufferOffset    uint32
	NtChallengeResponseLen             uint16
	NtChallengeResponseMaxLen          uint16
	NtChallengeResponseBufferOffset    uint32
	DomainNameLen                      uint16
	DomainNameMaxLen                   uint16
	DomainNameBufferOffset             uint32
	UserNameLen                        uint16
	UserNameMaxLen                     uint16
	UserNameBufferOffset               uint32
	WorkstationLen                     uint16
	WorkstationMaxLen                  uint16
	WorkstationBufferOffset            uint32
	EncryptedRandomSessionLen          uint16
	EncryptedRandomSessionMaxLen       uint16
	EncryptedRandomSessionBufferOffset uint16
	NegotiateFlags                     uint32
	Version                            NVersion
	MIC                                [16]byte
	Payload                            string
}

type NTLMv2 struct {
	domain              string
	user                string
	password            string
	negotiateMessage    *NegotiateMessage
	challengeMessage    *ChallengeMessage
	authenticateMessage *AuthenticateMessage
}

func NewNTLMv2(domain, user, password string) *NTLMv2 {
	return &NTLMv2{
		domain:   domain,
		user:     user,
		password: password,
	}
}

// generate first handshake messgae
func (n *NTLMv2) GetNegotiateMessage() *NegotiateMessage {
	n.negotiateMessage = &NegotiateMessage{
		NegotiateFlags: NTLMSSP_NEGOTIATE_KEY_EXCH |
			NTLMSSP_NEGOTIATE_128 |
			NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
			NTLMSSP_NEGOTIATE_NTLM |
			NTLMSSP_NEGOTIATE_SEAL |
			NTLMSSP_NEGOTIATE_SIGN |
			NTLMSSP_REQUEST_TARGET |
			NTLMSSP_NEGOTIATE_UNICODE,
	}
	return n.negotiateMessage
}

func (n *NTLMv2) GetAuthenticateMessage() *AuthenticateMessage {
	n.challengeMessage = &ChallengeMessage{}
	// todo
	return nil
}
