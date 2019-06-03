package lic

import (
	"github.com/icodeface/grdp/core"
	"github.com/icodeface/grdp/glog"
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

type LicensePacket struct {
	bMsgtype         uint8
	Flag             uint8
	wMsgSize         uint16
	LicensingMessage []byte
}

func readLicensePacket(r io.Reader) *LicensePacket {
	l := &LicensePacket{}
	l.bMsgtype, _ = core.ReadUInt8(r)
	l.Flag, _ = core.ReadUInt8(r)
	l.wMsgSize, _ = core.ReadUint16LE(r)
	l.LicensingMessage, _ = core.ReadBytes(int(l.wMsgSize-4), r)
	return l
}

func ReceiveLicensePacket(r io.Reader) bool {
	p := readLicensePacket(r)
	glog.Debug("ReceiveLicensePacket type is", p.bMsgtype)
	switch p.bMsgtype {
	case NEW_LICENSE:
		return true
	case LICENSE_REQUEST, PLATFORM_CHALLENGE:
		return false
	case ERROR_ALERT:
		glog.Info("ReceiveLicensePacket error alert")
		return true
	default:
		return false
	}
}
