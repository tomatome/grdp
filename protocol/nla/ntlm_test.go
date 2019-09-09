package nla_test

import (
	"bytes"
	"encoding/hex"
	"github.com/icodeface/grdp/protocol/nla"
	"github.com/lunixbochs/struc"
	"testing"
)

func TestNewNegotiateMessage(t *testing.T) {
	ntlm := nla.NewNTLMv2("", "", "")
	negoMsg := ntlm.GetNegotiateMessage()
	buff := &bytes.Buffer{}
	struc.Pack(buff, negoMsg)

	result := hex.EncodeToString(buff.Bytes())
	expected := "4e544c4d535350000100000035820860000000000000000000000000000000000000000000000000"

	if result != expected {
		t.Error(result, " not equals to", expected)
	}
}
