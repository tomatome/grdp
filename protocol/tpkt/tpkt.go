package tpkt

import (
	"github.com/chuckpreslar/emission"
	"net"
)

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
	t := &TPKT{*emission.NewEmitter(), conn, nil}
	return t
}

func (t *TPKT) Read(b []byte) (n int, err error) {
	return t.conn.Read(b)
}

func (t *TPKT) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (t *TPKT) Close() error {
	return t.conn.Close()
}

func (t *TPKT) recvHeader() {

}

func (t *TPKT) recvExtendedHeader() {

}

func (t *TPKT) recvData() {

}

func (t *TPKT) recvExtendedFastPathHeader() {

}

func (t *TPKT) recvFastPath() {

}
