package pdu

type MsgType uint16

const (
	CB_MONITOR_READY         = 0x0001
	CB_FORMAT_LIST           = 0x0002
	CB_FORMAT_LIST_RESPONSE  = 0x0003
	CB_FORMAT_DATA_REQUEST   = 0x0004
	CB_FORMAT_DATA_RESPONSE  = 0x0005
	CB_TEMP_DIRECTORY        = 0x0006
	CB_CLIP_CAPS             = 0x0007
	CB_FILECONTENTS_REQUEST  = 0x0008
	CB_FILECONTENTS_RESPONSE = 0x0009
	CB_LOCK_CLIPDATA         = 0x000A
	CB_UNLOCK_CLIPDATA       = 0x000B
)

type MsgFlags uint16

const (
	CB_RESPONSE_OK   = 0x0001
	CB_RESPONSE_FAIL = 0x0002
	CB_ASCII_NAMES   = 0x0004
)

type DwFlags uint32

const (
	FILECONTENTS_SIZE  = 0x00000001
	FILECONTENTS_RANGE = 0x00000002
)

type ClipboardPDUHeader struct {
	MsgType  uint16
	MsgFlags uint16
	DataLen  uint32
}

type ClipPDU struct {
	ClipPDUHeader *ClipboardPDUHeader
	Payload       []byte
	StreamId      uint32
	Lindex        int32
	DwFlags       uint32
	NPositionLow  uint32
	NPositionHigh uint32
	CbRequested   uint32
	ClipDataId    uint32
}

func NewClipPDU(data []byte) *ClipPDU {
	size := len(data)
	return &ClipPDU{
		ClipPDUHeader: &ClipboardPDUHeader{CB_FILECONTENTS_REQUEST, 0, uint32(size)},
		Payload:       data,
		DwFlags:       FILECONTENTS_SIZE,
		CbRequested:   0x00000008,
	}
}
