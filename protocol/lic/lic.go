package lic

import (
	"github.com/icodeface/grdp/core"
	"io"
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
