package tpkt

import (
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
	conn    net.Conn
	secFlag byte
}

func New(conn net.Conn) *TPKT {
	t := &TPKT{conn, nil}
	return t
}

func (t *TPKT) Read(b []byte) (n int, err error) {
	return t.conn.Read(b)
}

/**
 * Send message throught TPKT layer
 * @param message {type.*}
 */
func (t *TPKT) Write(b []byte) (n int, err error) {
	return 0, nil
}

/**
 * close stack
 */
func (t *TPKT) Close() error {
	return t.conn.Close()
}
