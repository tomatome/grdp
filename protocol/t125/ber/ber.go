package ber

import (
	"github.com/icodeface/grdp/core"
	"io"
)

const (
	CLASS_MASK uint8 = 0xC0
	CLASS_UNIV       = 0x00
	CLASS_APPL       = 0x40
	CLASS_CTXT       = 0x80
	CLASS_PRIV       = 0xC0
)

const (
	PC_MASK      uint8 = 0x20
	PC_PRIMITIVE       = 0x00
	PC_CONSTRUCT       = 0x20
)

const (
	TAG_MASK            uint8 = 0x1F
	TAG_BOOLEAN               = 0x01
	TAG_INTEGER               = 0x02
	TAG_BIT_STRING            = 0x03
	TAG_OCTET_STRING          = 0x04
	TAG_OBJECT_IDENFIER       = 0x06
	TAG_ENUMERATED            = 0x0A
	TAG_SEQUENCE              = 0x10
	TAG_SEQUENCE_OF           = 0x10
)

func berPC(pc bool) uint8 {
	if pc {
		return PC_CONSTRUCT
	}
	return PC_PRIMITIVE
}

func WriteUniversalTag(tag uint8, pc bool, w io.Writer) {
	core.WriteUInt8((CLASS_UNIV|berPC(pc))|(TAG_MASK&tag), w)
}

func WriteLength(size int, w io.Writer) {
	if size > 0x7f {
		core.WriteUInt8(0x82, w)
		core.WriteUInt16BE(uint16(size), w)
	} else {
		core.WriteUInt8(uint8(size), w)
	}
}

func WriteInteger(n int, w io.Writer) {
	WriteUniversalTag(TAG_INTEGER, false, w)
	if n <= 0xff {
		WriteLength(1, w)
		core.WriteUInt8(uint8(n), w)
	} else if n <= 0xffff {
		WriteLength(2, w)
		core.WriteUInt16BE(uint16(n), w)
	} else {
		WriteLength(4, w)
		core.WriteUInt32BE(uint32(n), w)
	}
}

func WriteOctetstring(str string, w io.Writer) {
	WriteUniversalTag(TAG_OCTET_STRING, false, w)
	WriteLength(len(str), w)
	core.WriteBytes([]byte(str), w)
}

func WriteBoolean(b bool, w io.Writer) {
	bb := uint8(0)
	if b {
		bb = uint8(0xff)
	}
	WriteUniversalTag(TAG_BOOLEAN, false, w)
	WriteLength(1, w)
	core.WriteUInt8(bb, w)
}

func WriteApplicationTag(tag uint8, size int, w io.Writer) {
	if tag > 30 {
		core.WriteUInt8((CLASS_APPL|PC_CONSTRUCT)|TAG_MASK, w)
		core.WriteUInt8(tag, w)
		WriteLength(size, w)
	} else {
		core.WriteUInt8((CLASS_APPL|PC_CONSTRUCT)|(TAG_MASK&tag), w)
		WriteLength(size, w)
	}
}

func WriteEncodedDomainParams(data []byte, w io.Writer) {
	WriteUniversalTag(TAG_SEQUENCE, true, w)
	WriteLength(len(data), w)
	core.WriteBytes(data, w)
}
