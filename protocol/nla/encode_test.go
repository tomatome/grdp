package nla_test

import (
	"encoding/hex"
	"github.com/icodeface/grdp/protocol/nla"
	"testing"
)

func TestNTOWFv2(t *testing.T) {
	res := hex.EncodeToString(nla.NTOWFv2("", "", ""))
	expected := "f4c1a15dd59d4da9bd595599220d971a"
	if res != expected {
		t.Error(res, "not equal to", expected)
	}

	res = hex.EncodeToString(nla.NTOWFv2("user", "pwd", "dom"))
	expected = "652feb8208b3a8a6264c9c5d5b820979"
	if res != expected {
		t.Error(res, "not equal to", expected)
	}

}
