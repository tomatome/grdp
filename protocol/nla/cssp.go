package nla

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type NegoToken []byte

type NegoData struct {
	Tokens []NegoToken
}

type TSRequest struct {
	Version    int
	NegoTokens *NegoData
	AuthInfo   string
	PubKeyAuth string
	ErrorCode  int
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

func EncodeDERTRequest(negoMsgs []*NegotiateMessage, authInfo string, pubKeyAuth string) []byte {
	buff := bytes.Buffer{}

	req := &TSRequest{
		Version: 2,
	}

	negoData := &NegoData{
		Tokens: make([]NegoToken, 0),
	}
	for _, msg := range negoMsgs {
		fmt.Println(hex.EncodeToString(msg.Serialize()))
		token := NegoToken(msg.Serialize())
		negoData.Tokens = append(negoData.Tokens, token)
	}

	if len(negoMsgs) > 0 {
		req.NegoTokens = negoData
	}

	if len(authInfo) > 0 {
		// todo
	}

	if len(pubKeyAuth) > 0 {
		// todo
	}

	fmt.Println(req)
	return buff.Bytes()

	//
	// 302fa003020102a12830263024a0220420
	// tokens:
	// 4e544c4d53535000010000003582086000000000000000000000000000000000
}

func DecodeDERTRequest(s []byte) (*TSRequest, error) {
	return nil, nil
}
