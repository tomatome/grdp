package nla

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"golang.org/x/crypto/md4"
)

func convertUTF16ToLittleEndianBytes(u []uint16) []byte {
	b := make([]byte, 2*len(u))
	for index, value := range u {
		binary.LittleEndian.PutUint16(b[index*2:], value)
	}
	return b
}

// s.encode('utf-16le')
func UnicodeEncode(p string) []byte {
	return convertUTF16ToLittleEndianBytes(utf16.Encode([]rune(p)))
}

func UnicodeDecode(p []byte) string {
	b := make([]byte, 2)
	n := make([]uint16, 0, len(p)/2)
	for i, v := range p {
		if i%2 == 0 {
			b[0] = v
		} else {
			b[1] = v
			a := binary.LittleEndian.Uint16(b)
			n = append(n, a)
		}
	}
	return string(utf16.Decode(n))
}

func MD4(data []byte) []byte {
	h := md4.New()
	h.Write(data)
	return h.Sum(nil)
}

func MD5(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func HMAC_MD5(key, data []byte) []byte {
	h := hmac.New(md5.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// Version 2 of NTLM hash function
func NTOWFv2(password, user, domain string) []byte {
	return HMAC_MD5(MD4(UnicodeEncode(password)), UnicodeEncode(strings.ToUpper(user)+domain))
}

// Same as NTOWFv2
func LMOWFv2(password, user, domain string) []byte {
	return NTOWFv2(password, user, domain)
}

func RC4K(key, src []byte) []byte {
	result := make([]byte, len(src))
	rc4obj, _ := rc4.NewCipher(key)
	rc4obj.XORKeyStream(result, src)
	return result
}
