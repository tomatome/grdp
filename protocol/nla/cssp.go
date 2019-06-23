package nla

import "bytes"

type NegoToken struct {
}

type NegoData struct {
}

type TSRequest struct {
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

	return buff.Bytes()
}

func DecodeDERTRequest(s []byte) (*TSRequest, error) {
	return nil, nil
}
