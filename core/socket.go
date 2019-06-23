package core

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/nla"
	"net"
)

type SocketLayer struct {
	conn    net.Conn
	tlsConn *tls.Conn
	ntlm    *nla.NTLMv2
}

func NewSocketLayer(conn net.Conn, ntlm *nla.NTLMv2) *SocketLayer {
	l := &SocketLayer{
		conn:    conn,
		tlsConn: nil,
		ntlm:    ntlm,
	}
	return l
}

func (s *SocketLayer) Read(b []byte) (n int, err error) {
	if s.tlsConn != nil {
		return s.tlsConn.Read(b)
	}
	return s.conn.Read(b)
}

func (s *SocketLayer) Write(b []byte) (n int, err error) {
	if s.tlsConn != nil {
		return s.tlsConn.Write(b)
	}
	return s.conn.Write(b)
}

func (s *SocketLayer) Close() error {
	if s.tlsConn != nil {
		err := s.tlsConn.Close()
		if err != nil {
			return err
		}
	}
	return s.conn.Close()
}

func (s *SocketLayer) StartTLS() error {
	glog.Info("StartTLS")
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	s.tlsConn = tls.Client(s.conn, config)
	return s.tlsConn.Handshake()
}

func (s *SocketLayer) StartNLA() error {
	glog.Debug("todo StartNLA")
	err := s.StartTLS()
	if err != nil {
		return err
	}
	_, err = s.Write(nla.EncodeDERTRequest([]*nla.NegotiateMessage{s.ntlm.GetNegotiateMessage()}, "", ""))
	if err != nil {
		return err
	}
	resp := make([]byte, 1024)
	_, err = s.Read(resp)
	if err != nil {
		return err
	}
	return s.recvChallenge(resp)
}

func (s *SocketLayer) recvChallenge(data []byte) error {
	glog.Debug("recvChallenge", hex.EncodeToString(data))
	req, err := nla.DecodeDERTRequest(data)
	if err != nil {
		return err
	}
	fmt.Println(req)
	// todo

	resp := make([]byte, 1024)
	_, err = s.Read(resp)
	if err != nil {
		return err
	}
	return s.recvPubKeyInc(resp)
}

func (s *SocketLayer) recvPubKeyInc(data []byte) error {
	// todo
	return nil
}
