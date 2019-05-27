package core

import (
	"encoding/binary"
	"io"
)

type ReadBytesComplete func(result []byte, err error)

func StartReadBytes(len int, r io.Reader, cb ReadBytesComplete) {
	b := make([]byte, len)
	go func() {
		_, err := io.ReadFull(r, b)
		cb(b, err)
	}()
}

func ReadBytes(len int, r io.Reader) ([]byte, error) {
	b := make([]byte, len)
	length, err := io.ReadFull(r, b)
	return b[:length], err
}

func ReadByte(r io.Reader) (byte, error) {
	b, err := ReadBytes(1, r)
	return b[0], err
}

func ReadUInt8(r io.Reader) (uint8, error) {
	b, err := ReadBytes(1, r)
	return uint8(b[0]), err
}

func ReadUint16LE(r io.Reader) (uint16, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return 0, nil
	}
	return binary.LittleEndian.Uint16(b), nil
}

func ReadUint16BE(r io.Reader) (uint16, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return 0, nil
	}
	return binary.BigEndian.Uint16(b), nil
}

func ReadUInt32LE(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return 0, nil
	}
	return binary.LittleEndian.Uint32(b), nil
}

func WriteByte(data byte, w io.Writer) (int, error) {
	b := make([]byte, 1)
	b[0] = byte(data)
	return w.Write(b)
}

func WriteBytes(data []byte, w io.Writer) (int, error) {
	return w.Write(data)
}

func WriteUInt8(data uint8, w io.Writer) (int, error) {
	b := make([]byte, 1)
	b[0] = byte(data)
	return w.Write(b)
}

func WriteUInt16BE(data uint16, w io.Writer) (int, error) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, data)
	return w.Write(b)
}

func WriteUInt16LE(data uint16, w io.Writer) (int, error) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, data)
	return w.Write(b)
}

func WriteUInt32LE(data uint32, w io.Writer) (int, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, data)
	return w.Write(b)
}

func WriteUInt32BE(data uint32, w io.Writer) (int, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, data)
	return w.Write(b)
}
