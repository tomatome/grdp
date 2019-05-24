package tpkt

import (
	"bytes"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/binary"
	"net"
)

// take idea from https://github.com/Madnikulin50/gordp

/**
 * Type of tpkt packet
 * Fastpath is use to shortcut RDP stack
 * @see http://msdn.microsoft.com/en-us/library/cc240621.aspx
 * @see http://msdn.microsoft.com/en-us/library/cc240589.aspx
 */
type TpktAction byte

const (
	FASTPATH_ACTION_FASTPATH TpktAction = 0x0
	FASTPATH_ACTION_X224                = 0x3
)

/**
 * TPKT layer of rdp stack
 */
type TPKT struct {
	emission.Emitter
	conn    net.Conn
	secFlag byte
}

func New(conn net.Conn) *TPKT {
	t := &TPKT{*emission.NewEmitter(), conn, 0}
	binary.StartReadBytes(2, conn, t.recvHeader)
	return t
}

func (t *TPKT) Read(b []byte) (n int, err error) {
	return t.conn.Read(b)
}

func (t *TPKT) Write(data []byte) (n int, err error) {
	buff := &bytes.Buffer{}
	binary.WriteUInt8(FASTPATH_ACTION_X224, buff)
	binary.WriteUInt8(0, buff)
	binary.WriteUInt16BE(uint16(len(data)+4), buff)
	buff.Write(data)
	fmt.Println("tpkt Write", buff.Bytes())
	return t.conn.Write(buff.Bytes())
}

func (t *TPKT) Close() error {
	return t.conn.Close()
}

func (t *TPKT) recvHeader(s []byte, err error) {
	fmt.Println("recvHeader", s, err)
	if err != nil {
		t.Emit("error", err)
		return
	}
	version := s[0]
	if version == FASTPATH_ACTION_X224 {
		binary.StartReadBytes(2, t.conn, t.recvExtendedHeader)
	} else {
		t.secFlag = (version >> 6) & 0x3
		length := int(s[1])
		if length&0x80 != 0 {
			binary.StartReadBytes(1, t.conn, func(s []byte, err error) {
				t.recvExtendedFastPathHeader(s, length, err)
			})
		} else {
			binary.StartReadBytes(length-2, t.conn, t.recvFastPath)
		}
	}
}

func (t *TPKT) recvExtendedHeader(s []byte, err error) {
	fmt.Println("recvExtendedHeader", s, err)
	if err != nil {
		return
	}
	r := bytes.NewReader(s)
	size, _ := binary.ReadUint16BE(r)
	binary.StartReadBytes(int(size-4), t.conn, t.recvData)
}

func (t *TPKT) recvData(s []byte, err error) {
	fmt.Println("recvData", s, err)
	if err != nil {
		return
	}
	t.Emit("data", s)
	binary.StartReadBytes(2, t.conn, t.recvHeader)
}

func (t *TPKT) recvExtendedFastPathHeader(s []byte, length int, err error) {
	fmt.Println("recvExtendedFastPathHeader", s, length, err)

}

func (t *TPKT) recvFastPath(s []byte, err error) {
	fmt.Println("recvFastPath", s, err)
}
