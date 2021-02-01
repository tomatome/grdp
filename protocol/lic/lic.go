package lic

import (
	"io"

	"github.com/tomatome/grdp/core"
)

const (
	LICENSE_REQUEST             = 0x01
	PLATFORM_CHALLENGE          = 0x02
	NEW_LICENSE                 = 0x03
	UPGRADE_LICENSE             = 0x04
	LICENSE_INFO                = 0x12
	NEW_LICENSE_REQUEST         = 0x13
	PLATFORM_CHALLENGE_RESPONSE = 0x15
	ERROR_ALERT                 = 0xFF
)

// error code
const (
	ERR_INVALID_SERVER_CERTIFICATE = 0x00000001
	ERR_NO_LICENSE                 = 0x00000002
	ERR_INVALID_SCOPE              = 0x00000004
	ERR_NO_LICENSE_SERVER          = 0x00000006
	STATUS_VALID_CLIENT            = 0x00000007
	ERR_INVALID_CLIENT             = 0x00000008
	ERR_INVALID_PRODUCTID          = 0x0000000B
	ERR_INVALID_MESSAGE_LEN        = 0x0000000C
	ERR_INVALID_MAC                = 0x00000003
)

// state transition
const (
	ST_TOTAL_ABORT          = 0x00000001
	ST_NO_TRANSITION        = 0x00000002
	ST_RESET_PHASE_TO_START = 0x00000003
	ST_RESEND_LAST_MESSAGE  = 0x00000004
)

type ErrorMessage struct {
	DwErrorCode        uint32
	DwStateTransaction uint32
	Blob               []byte
}

func readErrorMessage(r io.Reader) *ErrorMessage {
	m := &ErrorMessage{}
	m.DwErrorCode, _ = core.ReadUInt32LE(r)
	m.DwStateTransaction, _ = core.ReadUInt32LE(r)
	return m
}

type LicensePacket struct {
	BMsgtype         uint8
	Flag             uint8
	WMsgSize         uint16
	LicensingMessage interface{}
}

func ReadLicensePacket(r io.Reader) *LicensePacket {
	l := &LicensePacket{}
	l.BMsgtype, _ = core.ReadUInt8(r)
	l.Flag, _ = core.ReadUInt8(r)
	l.WMsgSize, _ = core.ReadUint16LE(r)

	switch l.BMsgtype {
	case ERROR_ALERT:
		l.LicensingMessage = readErrorMessage(r)
	default:
		l.LicensingMessage, _ = core.ReadBytes(int(l.WMsgSize-4), r)
	}
	return l
}

/*
@summary:  Send by server to signal license request
            server -> client
@see: http://msdn.microsoft.com/en-us/library/cc241914.aspx
*/
type ServerLicenseRequest struct {
	/*ServerRandom []byte
	  ProductInfo = ProductInformation()
	  KeyExchangeList = LicenseBinaryBlob(BinaryBlobType.BB_KEY_EXCHG_ALG_BLOB)
	  ServerCertificate = LicenseBinaryBlob(BinaryBlobType.BB_CERTIFICATE_BLOB)
	  ScopeList = ScopeList()*/
}

/*
@summary:  Send by client to ask new license for client.
            RDPY doesn'support license reuse, need it in futur version
@see: http://msdn.microsoft.com/en-us/library/cc241918.aspx
 	#RSA and must be only RSA
    #pure microsoft client ;-)
    #http://msdn.microsoft.com/en-us/library/1040af38-c733-4fb3-acd1-8db8cc979eda#id10
*/
type ClientNewLicenseRequest struct {
	/*PreferredKeyExchangeAlg uint32

	  PlatformId uint32
	  ClientRandom []byte
	  EncryptedPreMasterSecret = LicenseBinaryBlob(BinaryBlobType.BB_RANDOM_BLOB)
	  ClientUserName = LicenseBinaryBlob(BinaryBlobType.BB_CLIENT_USER_NAME_BLOB)
	  ClientMachineName = LicenseBinaryBlob(BinaryBlobType.BB_CLIENT_MACHINE_NAME_BLOB)*/
}

/*
@summary: challenge send from server to client
@see: http://msdn.microsoft.com/en-us/library/cc241921.aspx
*/
type ServerPlatformChallenge struct {

	/*ConnectFlags uint32
	  EncryptedPlatformChallenge = LicenseBinaryBlob(BinaryBlobType.BB_ANY_BLOB)
	  MACData [16]byte*/
}

func (l *LicensePacket) Serialize() []byte {
	//todo
	return nil
}
