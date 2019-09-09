package core

import (
	"encoding/hex"
	"fmt"
	"github.com/icodeface/grdp/glog"
	"github.com/icodeface/grdp/protocol/nla"
	"github.com/icodeface/tls"
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
		InsecureSkipVerify:       true,
		MinVersion:               tls.VersionTLS10,
		MaxVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
	}
	s.tlsConn = tls.Client(s.conn, config)
	return s.tlsConn.Handshake()
}

func (s *SocketLayer) StartNLA() error {
	glog.Info("StartNLA")
	err := s.StartTLS()
	if err != nil {
		glog.Info("start tls failed", err)
		return err
	}
	req := nla.EncodeDERTRequest([]nla.Message{s.ntlm.GetNegotiateMessage()}, "", "")
	_, err = s.Write(req)
	if err != nil {
		glog.Info("send NegotiateMessage", err)
		return err
	}

	resp := make([]byte, 1024)
	n, err := s.Read(resp)
	if err != nil {
		return fmt.Errorf("read %s", err)
	}
	return s.recvChallenge(resp[:n])
}

func (s *SocketLayer) recvChallenge(data []byte) error {
	glog.Debug("recvChallenge", hex.EncodeToString(data))
	tsreq, err := nla.DecodeDERTRequest(data)
	if err != nil {
		return err
	}

	// get pubkey

	pubkey := ""

	msg := s.ntlm.GetAuthenticateMessage(tsreq.NegoTokens[0].Data)
	req := nla.EncodeDERTRequest([]nla.Message{msg}, "", pubkey)
	_, err = s.Write(req)
	if err != nil {
		glog.Info("send AuthenticateMessage", err)
		return err
	}

	//resp := make([]byte, 1024)
	//_, err = s.Read(resp)
	//if err != nil {
	//	return err
	//}
	//return s.recvPubKeyInc(resp)

	fmt.Println("todo")
	return nil

}

func (s *SocketLayer) recvPubKeyInc(data []byte) error {
	// todo
	return nil
}
