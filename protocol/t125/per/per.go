package per

import (
	"bytes"
	"github.com/icodeface/grdp/core"
	"io"
)

func ReadEnumerates(r io.Reader) (uint8, error) {
	return core.ReadUInt8(r)
}

func WriteInteger(n int, w io.Writer) {
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

func ReadInteger16(r io.Reader) (uint16, error) {
	return core.ReadUint16BE(r)
}

func WriteInteger16(value uint16, w io.Writer) {
	core.WriteUInt16BE(value, w)
}

/**
 * @param choice {integer}
 * @returns {type.UInt8} choice per encoded
 */
func WriteChoice(choice uint8, w io.Writer) {
	core.WriteUInt8(choice, w)
}

/**
 * @param value {raw} value to convert to per format
 * @returns type objects per encoding value
 */
func WriteLength(value int, w io.Writer) {
	if value > 0x7f {
		core.WriteUInt16BE(uint16(value|0x8000), w)
	} else {
		core.WriteUInt8(uint8(value), w)
	}
}

func ReadLength(r io.Reader) (uint16, error) {
	b, err := core.ReadUInt8(r)
	if err != nil {
		return 0, nil
	}
	var size uint16
	if b&0x80 > 0 {
		b = b &^ 0x80
		size = uint16(b) << 8
		left, _ := core.ReadUInt8(r)
		size += uint16(left)
	} else {
		size = uint16(b)
	}
	return size, nil
}

/**
 * @param oid {array} oid to write
 * @returns {type.Component} per encoded object identifier
 */
func WriteObjectIdentifier(oid []byte, w io.Writer) {
	core.WriteUInt8(5, w)
	core.WriteByte((oid[0]<<4)&(oid[1]&0x0f), w)
	core.WriteByte(oid[2], w)
	core.WriteByte(oid[3], w)
	core.WriteByte(oid[4], w)
	core.WriteByte(oid[5], w)
}

/**
 * @param selection {integer}
 * @returns {type.UInt8} per encoded selection
 */
func WriteSelection(selection uint8, w io.Writer) {
	core.WriteUInt8(selection, w)
}

func WriteNumericString(s string, minValue int, w io.Writer) {
	length := len(s)
	mLength := minValue
	if length >= minValue {
		mLength = length - minValue
	}
	buff := &bytes.Buffer{}
	for i := 0; i < length; i += 2 {
		c1 := int(s[i])
		c2 := 0x30
		if i+1 < length {
			c2 = int(s[i+1])
		}
		c1 = (c1 - 0x30) % 10
		c2 = (c2 - 0x30) % 10
		core.WriteUInt8(uint8((c1<<4)|c2), buff)
	}
	WriteLength(mLength, w)
	w.Write(buff.Bytes())
}

func WritePadding(length int, w io.Writer) {
	b := make([]byte, length)
	w.Write(b)
}

func WriteNumberOfSet(n int, w io.Writer) {
	core.WriteUInt8(uint8(n), w)
}

/**
 * @param oStr {String}
 * @param minValue {integer} default 0
 * @returns {type.Component} per encoded octet stream
 */
func WriteOctetStream(oStr string, minValue int, w io.Writer) {
	length := len(oStr)
	mlength := minValue

	if length-minValue >= 0 {
		mlength = length - minValue
	}
	WriteLength(mlength, w)
	w.Write([]byte(oStr)[:length])
}
