package nla

import (
	"encoding/asn1"

	"github.com/tomatome/grdp/glog"
)

type NegoToken struct {
	Data []byte `asn1:"explicit,tag:0"`
}

type TSRequest struct {
	Version    int         `asn1:"explicit,tag:0"`
	NegoTokens []NegoToken `asn1:"optional,explicit,tag:1"`
	AuthInfo   string      `asn1:"optional,explicit,tag:2"`
	PubKeyAuth string      `asn1:"optional,explicit,tag:3"`
	ErrorCode  int         `asn1:"optional,explicit,tag:4"`
}

type TSCredentials struct {
	CredType    int    `asn1:"explicit,tag:0"`
	Credentials string `asn1:"explicit,tag:1"`
}

type TSPasswordCreds struct {
	DomainName string `asn1:"explicit,tag:0"`
	UserName   string `asn1:"explicit,tag:1"`
	Password   string `asn1:"explicit,tag:2"`
}

type TSCspDataDetail struct {
	KeySpec       int    `asn1:"explicit,tag:0"`
	CardName      string `asn1:"explicit,tag:1"`
	ReaderName    string `asn1:"explicit,tag:2"`
	ContainerName string `asn1:"explicit,tag:3"`
	CspName       string `asn1:"explicit,tag:4"`
}

type TSSmartCardCreds struct {
	Pin        string            `asn1:"explicit,tag:0"`
	CspData    []TSCspDataDetail `asn1:"explicit,tag:1"`
	UserHint   string            `asn1:"explicit,tag:2"`
	DomainHint string            `asn1:"explicit,tag:3"`
}

type OpenSSLRSAPublicKey struct {
}

func EncodeDERTRequest(msgs []Message, authInfo []byte, pubKeyAuth []byte) []byte {
	req := TSRequest{
		Version:    2,
		NegoTokens: make([]NegoToken, 0),
	}

	for _, msg := range msgs {
		token := NegoToken{msg.Serialize()}
		req.NegoTokens = append(req.NegoTokens, token)
	}

	if len(authInfo) > 0 {
		// todo
	}

	if len(pubKeyAuth) > 0 {
		// todo
	}

	result, err := asn1.Marshal(req)
	if err != nil {
		glog.Error(err)
	}
	return result
}

func DecodeDERTRequest(s []byte) (*TSRequest, error) {
	treq := &TSRequest{}
	_, err := asn1.Unmarshal(s, treq)
	return treq, err
}
func EncodeDERTCredentials(domain, username, password string) []byte {
	tpas := TSPasswordCreds{domain, username, password}
	result, err := asn1.Marshal(tpas)
	if err != nil {
		glog.Error(err)
	}
	tcre := TSCredentials{1, string(result)}
	result, err = asn1.Marshal(tcre)
	if err != nil {
		glog.Error(err)
	}

	return result
}

func DecodeDERTCredentials(s []byte) (*TSCredentials, error) {
	tcre := &TSCredentials{}
	_, err := asn1.Unmarshal(s, tcre)
	return tcre, err
}
