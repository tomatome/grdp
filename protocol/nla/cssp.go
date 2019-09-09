package nla

import (
	"encoding/asn1"
	"github.com/icodeface/grdp/glog"
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
}

type TSPasswordCreds struct {
}

type TSCspDataDetail struct {
}

type TSSmartCardCreds struct {
}

type OpenSSLRSAPublicKey struct {
}

func EncodeDERTRequest(msgs []Message, authInfo string, pubKeyAuth string) []byte {
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
