package pdu

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tomatome/grdp/glog"

	"github.com/tomatome/grdp/core"
)

type ControlFlag uint8

const (
	TS_STANDARD             = 0x01
	TS_SECONDARY            = 0x02
	TS_BOUNDS               = 0x04
	TS_TYPE_CHANGE          = 0x08
	TS_DELTA_COORDINATES    = 0x10
	TS_ZERO_BOUNDS_DELTAS   = 0x20
	TS_ZERO_FIELD_BYTE_BIT0 = 0x40
	TS_ZERO_FIELD_BYTE_BIT1 = 0x80
)

type PrimaryOrderType uint8

const (
	ORDER_TYPE_DSTBLT             = 0x00 //0
	ORDER_TYPE_PATBLT             = 0x01 //1
	ORDER_TYPE_SCRBLT             = 0x02 //2
	ORDER_TYPE_DRAWNINEGRID       = 0x07 //7
	ORDER_TYPE_MULTI_DRAWNINEGRID = 0x08 //8
	ORDER_TYPE_LINETO             = 0x09 //9
	ORDER_TYPE_OPAQUERECT         = 0x0A //10
	ORDER_TYPE_SAVEBITMAP         = 0x0B //11
	ORDER_TYPE_MEMBLT             = 0x0D //13
	ORDER_TYPE_MEM3BLT            = 0x0E //14
	ORDER_TYPE_MULTIDSTBLT        = 0x0F //15
	ORDER_TYPE_MULTIPATBLT        = 0x10 //16
	ORDER_TYPE_MULTISCRBLT        = 0x11 //17
	ORDER_TYPE_MULTIOPAQUERECT    = 0x12 //18
	ORDER_TYPE_FAST_INDEX         = 0x13 //19
	ORDER_TYPE_POLYGON_SC         = 0x14 //20
	ORDER_TYPE_POLYGON_CB         = 0x15 //21
	ORDER_TYPE_POLYLINE           = 0x16 //22
	ORDER_TYPE_FAST_GLYPH         = 0x18 //24
	ORDER_TYPE_ELLIPSE_SC         = 0x19 //25
	ORDER_TYPE_ELLIPSE_CB         = 0x1A //26
	ORDER_TYPE_GLYPHINDEX         = 0x1B //27
)

type SecondaryOrderType uint8

const (
	ORDER_TYPE_BITMAP_UNCOMPRESSED     = 0x00
	ORDER_TYPE_CACHE_COLOR_TABLE       = 0x01
	ORDER_TYPE_CACHE_BITMAP_COMPRESSED = 0x02
	ORDER_TYPE_CACHE_GLYPH             = 0x03
	ORDER_TYPE_BITMAP_UNCOMPRESSED_V2  = 0x04
	ORDER_TYPE_BITMAP_COMPRESSED_V2    = 0x05
	ORDER_TYPE_CACHE_BRUSH             = 0x07
	ORDER_TYPE_BITMAP_COMPRESSED_V3    = 0x08
)

func (s SecondaryOrderType) String() string {
	name := "Unknown"
	switch s {
	case ORDER_TYPE_BITMAP_UNCOMPRESSED:
		name = "Cache Bitmap"
	case ORDER_TYPE_CACHE_COLOR_TABLE:
		name = "Cache Color Table"
	case ORDER_TYPE_CACHE_BITMAP_COMPRESSED:
		name = "Cache Bitmap (Compressed)"
	case ORDER_TYPE_CACHE_GLYPH:
		name = "Cache Glyph"
	case ORDER_TYPE_BITMAP_UNCOMPRESSED_V2:
		name = "Cache Bitmap V2"
	case ORDER_TYPE_BITMAP_COMPRESSED_V2:
		name = "Cache Bitmap V2 (Compressed)"
	case ORDER_TYPE_CACHE_BRUSH:
		name = "Cache Brush"
	case ORDER_TYPE_BITMAP_COMPRESSED_V3:
		name = "Cache Bitmap V3"
	}
	return fmt.Sprintf("[0x%02d] %s", s, name)
}

/* Alternate Secondary Drawing Orders */
const (
	ORDER_TYPE_SWITCH_SURFACE          = 0x00
	ORDER_TYPE_CREATE_OFFSCREEN_BITMAP = 0x01
	ORDER_TYPE_STREAM_BITMAP_FIRST     = 0x02
	ORDER_TYPE_STREAM_BITMAP_NEXT      = 0x03
	ORDER_TYPE_CREATE_NINE_GRID_BITMAP = 0x04
	ORDER_TYPE_GDIPLUS_FIRST           = 0x05
	ORDER_TYPE_GDIPLUS_NEXT            = 0x06
	ORDER_TYPE_GDIPLUS_END             = 0x07
	ORDER_TYPE_GDIPLUS_CACHE_FIRST     = 0x08
	ORDER_TYPE_GDIPLUS_CACHE_NEXT      = 0x09
	ORDER_TYPE_GDIPLUS_CACHE_END       = 0x0A
	ORDER_TYPE_WINDOW                  = 0x0B
	ORDER_TYPE_COMPDESK_FIRST          = 0x0C
	ORDER_TYPE_FRAME_MARKER            = 0x0D
)

const (
	GLYPH_FRAGMENT_NOP = 0x00
	GLYPH_FRAGMENT_USE = 0xFE
	GLYPH_FRAGMENT_ADD = 0xFF

	CBR2_HEIGHT_SAME_AS_WIDTH      = 0x01
	CBR2_PERSISTENT_KEY_PRESENT    = 0x02
	CBR2_NO_BITMAP_COMPRESSION_HDR = 0x08
	CBR2_DO_NOT_CACHE              = 0x10
)

type FastPathOrdersPDU struct {
	NumberOrders uint16
	OrderPdus    []OrderPdu
}

type OrderPdu struct {
	ControlFlags uint8
	Data         OrderData
}

const (
	ORDER_PRIMARY = iota + 1
	ORDER_SECONDARY
	ORDER_ALTSEC
)

type OrderData interface {
	Type() int
}

type orderMemblt struct {
	colorTable uint8
	cacheId    uint8
	x          uint16
	y          uint16
	cx         uint16
	cy         uint16
	opcode     uint8
	srcx       int16
	srcy       int16
	cacheIdx   uint16
}

type Secondary struct {
}

func (*Secondary) Type() int {
	return ORDER_SECONDARY
}

type Primary struct {
	delta   bool
	present uint32
	Scrblt  Scrblt
}

func (*Primary) Type() int {
	return ORDER_PRIMARY
}
func (*FastPathOrdersPDU) FastPathUpdateType() uint8 {
	return FASTPATH_UPDATETYPE_ORDERS
}

func (f *FastPathOrdersPDU) Unpack(r io.Reader) error {
	f.NumberOrders, _ = core.ReadUint16LE(r)
	//glog.Info("NumberOrders:", f.NumberOrders)
	for i := 0; i < int(f.NumberOrders); i++ {
		var o OrderPdu
		cflags, _ := core.ReadUInt8(r)
		if cflags&TS_STANDARD == 0 {
			//glog.Info("Altsec order")
			o.processAltsecOrder(r, cflags)
			//return errors.New("Not support")
		} else if cflags&TS_SECONDARY != 0 {
			//glog.Info("Secondary order")
			o.processSecondaryOrder(r, cflags)
		} else {
			glog.Info("Primary order")
			o.processPrimaryOrder(r, cflags)
		}

		if f.OrderPdus == nil {
			f.OrderPdus = make([]OrderPdu, 0, f.NumberOrders)
		}
		f.OrderPdus = append(f.OrderPdus, o)
	}
	return nil
}
func (o *OrderPdu) processAltsecOrder(r io.Reader, cflags uint8) error {
	orderType := cflags >> 2
	//glog.Info("Altsec:", orderType)
	switch orderType {
	case ORDER_TYPE_SWITCH_SURFACE:
	case ORDER_TYPE_CREATE_OFFSCREEN_BITMAP:
	case ORDER_TYPE_STREAM_BITMAP_FIRST:
	case ORDER_TYPE_STREAM_BITMAP_NEXT:
	case ORDER_TYPE_CREATE_NINE_GRID_BITMAP:
	case ORDER_TYPE_GDIPLUS_FIRST:
	case ORDER_TYPE_GDIPLUS_NEXT:
	case ORDER_TYPE_GDIPLUS_END:
	case ORDER_TYPE_GDIPLUS_CACHE_FIRST:
	case ORDER_TYPE_GDIPLUS_CACHE_NEXT:
	case ORDER_TYPE_GDIPLUS_CACHE_END:
	case ORDER_TYPE_WINDOW:
	case ORDER_TYPE_COMPDESK_FIRST:
	case ORDER_TYPE_FRAME_MARKER:
		core.ReadUInt32LE(r)
	}

	return nil
}
func (o *OrderPdu) processSecondaryOrder(r io.Reader, cflags uint8) error {
	var sec Secondary
	length, _ := core.ReadUint16LE(r)
	flags, _ := core.ReadUint16LE(r)
	orderType, _ := core.ReadUInt8(r)

	glog.Info("Secondary:", SecondaryOrderType(orderType))

	b, _ := core.ReadBytes(int(length)+13-6, r)
	r0 := bytes.NewReader(b)

	switch orderType {
	case ORDER_TYPE_BITMAP_UNCOMPRESSED:
		fallthrough
	case ORDER_TYPE_CACHE_BITMAP_COMPRESSED:
		compressed := (orderType == ORDER_TYPE_CACHE_BITMAP_COMPRESSED)
		sec.updateCacheBitmapOrder(r0, compressed, flags)
	case ORDER_TYPE_BITMAP_UNCOMPRESSED_V2:
		fallthrough
	case ORDER_TYPE_BITMAP_COMPRESSED_V2:
		compressed := (orderType == ORDER_TYPE_BITMAP_COMPRESSED_V2)
		sec.updateCacheBitmapV2Order(r0, compressed, flags)
	case ORDER_TYPE_BITMAP_COMPRESSED_V3:
		sec.updateCacheBitmapV3Order(r0, flags)
	case ORDER_TYPE_CACHE_COLOR_TABLE:
		sec.updateCacheColorTableOrder(r0, flags)
	case ORDER_TYPE_CACHE_GLYPH:
		sec.updateCacheGlyphOrder(r0, flags)
	case ORDER_TYPE_CACHE_BRUSH:
		sec.updateCacheBrushOrder(r0, flags)
	default:
		glog.Debugf("Unsupport order type 0x%x", orderType)
	}

	return nil
}
func update_parse_bounds(r io.Reader) *Bounds {
	var bounds Bounds

	present, _ := core.ReadUInt8(r)

	if present&1 != 0 {
		readOrderCoord(r, &bounds.left, false)
	} else if present&16 != 0 {
		readOrderCoord(r, &bounds.left, true)
	}

	if present&2 != 0 {
		readOrderCoord(r, &bounds.top, false)
	} else if present&32 != 0 {
		readOrderCoord(r, &bounds.top, true)
	}

	if present&4 != 0 {
		readOrderCoord(r, &bounds.right, false)
	} else if present&64 != 0 {
		readOrderCoord(r, &bounds.right, true)
	}
	if present&8 != 0 {
		readOrderCoord(r, &bounds.bottom, false)
	} else if present&128 != 0 {
		readOrderCoord(r, &bounds.bottom, true)
	}

	return &bounds
}

var orderType uint8

func (o *OrderPdu) processPrimaryOrder(r io.Reader, cflags uint8) error {
	//var orderType uint8
	if cflags&TS_TYPE_CHANGE != 0 {
		glog.Info("ORDER_TYPE_CHANGE")
		orderType, _ = core.ReadUInt8(r)
		glog.Infof("orderType:0x%x", orderType)
	}
	size := 1
	switch orderType {
	case ORDER_TYPE_MEM3BLT, ORDER_TYPE_GLYPHINDEX:
		size = 3

	case ORDER_TYPE_PATBLT, ORDER_TYPE_MEMBLT, ORDER_TYPE_LINETO, ORDER_TYPE_POLYGON_CB, ORDER_TYPE_ELLIPSE_CB:
		size = 2
	}

	if cflags&TS_ZERO_FIELD_BYTE_BIT0 != 0 {
		size--
	}
	if cflags&TS_ZERO_FIELD_BYTE_BIT1 != 0 {
		if size < 2 {
			size = 0
		} else {
			size -= 2
		}
	}
	var present uint32
	for i := 0; i < size; i++ {
		bits, _ := core.ReadUInt8(r)
		present |= uint32(bits) << (i * 8)
	}

	if cflags&TS_BOUNDS != 0 {
		if cflags&TS_ZERO_BOUNDS_DELTAS == 0 {
			bounds := update_parse_bounds(r)
			glog.Debug(bounds)
		}
		//update ui
	}
	var p Primary
	p.delta = cflags&TS_DELTA_COORDINATES != 0
	p.present = present
	o.Data = &p

	glog.Info("Primary:", orderType)

	switch orderType {
	case ORDER_TYPE_DSTBLT:
		p.updateReadDstbltOrder(r)

	case ORDER_TYPE_PATBLT:
		p.updateReadPatbltOrder(r)

	case ORDER_TYPE_SCRBLT:
		p.updateReadScrbltOrder(r)

	case ORDER_TYPE_DRAWNINEGRID:
		//p.update_read_draw_nine_grid_order(r, orderInfo)

	case ORDER_TYPE_MULTI_DRAWNINEGRID:
		//p.update_read_multi_draw_nine_grid_order(r, orderInfo)

	case ORDER_TYPE_LINETO:
		//p.update_read_line_to_order(r, orderInfo)

	case ORDER_TYPE_OPAQUERECT:
		p.updateReadOpaqueRectOrder(r)

	case ORDER_TYPE_SAVEBITMAP:
		//p.update_read_save_bitmap_order(r, orderInfo)

	case ORDER_TYPE_MEMBLT:
		//p.update_read_memblt_order(r, orderInfo)

	case ORDER_TYPE_MEM3BLT:
		//p.update_read_mem3blt_order(r, orderInfo)

	case ORDER_TYPE_MULTIDSTBLT:
		//p.update_read_multi_dstblt_order(r, orderInfo)

	case ORDER_TYPE_MULTIPATBLT:
		//p.update_read_multi_patblt_order(r, orderInfo)

	case ORDER_TYPE_MULTISCRBLT:
		//p.update_read_multi_scrblt_order(r, orderInfo)

	case ORDER_TYPE_MULTIOPAQUERECT:
		//p.update_read_multi_opaque_rect_order(r, orderInfo)

	case ORDER_TYPE_FAST_INDEX:
		//p.update_read_fast_index_order(r, orderInfo)

	case ORDER_TYPE_POLYGON_SC:
		//p.update_read_polygon_sc_order(r, orderInfo)

	case ORDER_TYPE_POLYGON_CB:
		//p.update_read_polygon_cb_order(r, orderInfo)

	case ORDER_TYPE_POLYLINE:
		//p.update_read_polyline_order(r, orderInfo)

	case ORDER_TYPE_FAST_GLYPH:
		//p.update_read_fast_glyph_order(r, orderInfo)

	case ORDER_TYPE_ELLIPSE_SC:
		//p.update_read_ellipse_sc_order(r, orderInfo)

	case ORDER_TYPE_ELLIPSE_CB:
		//p.update_read_ellipse_cb_order(r, orderInfo)

	case ORDER_TYPE_GLYPHINDEX:
		//p.update_read_glyph_index_order(r, orderInfo)
	}

	return nil
}
func readOrderCoord(r io.Reader, coord *int32, delta bool) {
	if delta {
		change, _ := core.ReadUInt8(r)
		*coord += int32(change)
	} else {
		change, _ := core.ReadUint16LE(r)
		*coord = int32(change)
	}
}

type Dstblt struct {
	x      int32
	y      int32
	cx     int32
	cy     int32
	opcode uint8
}

func (p *Primary) updateReadDstbltOrder(r io.Reader) {
	glog.Infof("Dstblt Order")
	var d Dstblt
	if p.present&0x01 != 0 {
		readOrderCoord(r, &d.x, p.delta)
	}
	if p.present&0x02 != 0 {
		readOrderCoord(r, &d.y, p.delta)
	}
	if p.present&0x04 != 0 {
		readOrderCoord(r, &d.cx, p.delta)
	}
	if p.present&0x08 != 0 {
		readOrderCoord(r, &d.cy, p.delta)
	}
	if p.present&0x10 != 0 {
		d.opcode, _ = core.ReadUInt8(r)
	}
}

type Patblt struct {
	x        int32
	y        int32
	cx       int32
	cy       int32
	opcode   uint8
	bgcolour [4]uint8
	fgcolour [4]uint8
	brush    *Brush
}
type Brush struct {
	x     uint8
	y     uint8
	bpp   uint8
	style uint8
	hatch uint8
	index uint8
	data  []byte
}

func readOrderbrush(r io.Reader, present uint32) *Brush {
	var b Brush
	if present&1 != 0 {
		b.x, _ = core.ReadUInt8(r)
	}

	if present&2 != 0 {
		b.y, _ = core.ReadUInt8(r)
	}

	if present&4 != 0 {
		b.style, _ = core.ReadUInt8(r)
	}

	if present&8 != 0 {
		b.hatch, _ = core.ReadUInt8(r)
	}

	if present&16 != 0 {
		data, _ := core.ReadBytes(7, r)
		b.data = make([]byte, 0, 8)
		b.data = append(b.data, b.hatch)
		b.data = append(b.data, data...)
	}

	return &b
}
func (p *Primary) updateReadPatbltOrder(r io.Reader) {
	glog.Infof("Patblt Order")
	var d Patblt
	if p.present&0x01 != 0 {
		readOrderCoord(r, &d.x, p.delta)
	}
	if p.present&0x02 != 0 {
		readOrderCoord(r, &d.y, p.delta)
	}
	if p.present&0x04 != 0 {
		readOrderCoord(r, &d.cx, p.delta)
	}
	if p.present&0x08 != 0 {
		readOrderCoord(r, &d.cy, p.delta)
	}
	if p.present&0x10 != 0 {
		d.opcode, _ = core.ReadUInt8(r)
	}
	if p.present&0x0020 != 0 {
		b, g, r, a := updateReadColorRef(r)
		d.bgcolour[0], d.bgcolour[1], d.bgcolour[2], d.bgcolour[3] = b, g, r, a
	}
	if p.present&0x0040 != 0 {
		b, g, r, a := updateReadColorRef(r)
		d.fgcolour[0], d.fgcolour[1], d.fgcolour[2], d.fgcolour[3] = b, g, r, a
	}
	d.brush = readOrderbrush(r, p.present>>7)
}

type Scrblt struct {
	X      int32
	Y      int32
	Cx     int32
	Cy     int32
	Opcode uint8
	Srcx   int32
	Srcy   int32
}

func (p *Primary) updateReadScrbltOrder(r io.Reader) {
	glog.Infof("Scrblt Order")
	var d Scrblt
	if p.present&0x0001 != 0 {
		readOrderCoord(r, &d.X, p.delta)
	}
	if p.present&0x0002 != 0 {
		readOrderCoord(r, &d.Y, p.delta)
	}
	if p.present&0x0004 != 0 {
		readOrderCoord(r, &d.Cx, p.delta)
	}
	if p.present&0x0008 != 0 {
		readOrderCoord(r, &d.Cy, p.delta)
	}
	if p.present&0x0010 != 0 {
		d.Opcode, _ = core.ReadUInt8(r)
	}
	if p.present&0x0020 != 0 {
		readOrderCoord(r, &d.Srcx, p.delta)
	}
	if p.present&0x0040 != 0 {
		readOrderCoord(r, &d.Srcy, p.delta)
	}

	p.Scrblt = d
}

func (p *Primary) updateReadOpaqueRectOrder(r io.Reader) {
	glog.Infof("Opaque Rect Order")
}

/*Secondary*/
func (s *Secondary) updateCacheBitmapOrder(r io.Reader, compressed bool, flags uint16) {
	var cb CacheBitmapOrder
	cb.cacheId, _ = core.ReadUInt8(r)
	core.ReadUInt8(r)
	cb.bitmapWidth, _ = core.ReadUInt8(r)
	cb.bitmapHeight, _ = core.ReadUInt8(r)
	cb.bitmapBpp, _ = core.ReadUInt8(r)
	bitmapLength, _ := core.ReadUint16LE(r)
	cb.cacheIndex, _ = core.ReadUint16LE(r)
	var bitmapComprHdr []byte
	if compressed {
		if (flags & NO_BITMAP_COMPRESSION_HDR) == 0 {
			bitmapComprHdr, _ = core.ReadBytes(8, r)
			bitmapLength -= 8
		}
	}
	cb.bitmapComprHdr = bitmapComprHdr
	cb.bitmapDataStream, _ = core.ReadBytes(int(bitmapLength), r)
	cb.bitmapLength = bitmapLength

}

type CacheBitmapOrder struct {
	cacheId          uint8
	bitmapBpp        uint8
	bitmapWidth      uint8
	bitmapHeight     uint8
	bitmapLength     uint16
	cacheIndex       uint16
	bitmapComprHdr   []byte
	bitmapDataStream []byte
}

func getCbV2Bpp(bpp uint32) (b uint32) {
	switch bpp {
	case 3:
		b = 8
	case 4:
		b = 16
	case 5:
		b = 24
	case 6:
		b = 32
	default:
		b = 0
	}
	return
}

type CacheBitmapV2Order struct {
	cacheId            uint32
	flags              uint32
	key1               uint32
	key2               uint32
	bitmapBpp          uint32
	bitmapWidth        uint8
	bitmapHeight       uint8
	bitmapLength       uint16
	cacheIndex         uint32
	compressed         bool
	cbCompFirstRowSize uint16
	cbCompMainBodySize uint16
	cbScanWidth        uint16
	cbUncompressedSize uint16
	bitmapDataStream   []byte
}

func (s *Secondary) updateCacheBitmapV2Order(r io.Reader, compressed bool, flags uint16) {
	var cb CacheBitmapV2Order
	cb.cacheId = uint32(flags) & 0x0003
	cb.flags = (uint32(flags) & 0xFF80) >> 7
	bitsPerPixelId := (uint32(flags) & 0x0078) >> 3
	cb.bitmapBpp = getCbV2Bpp(bitsPerPixelId)

	if cb.flags&CBR2_PERSISTENT_KEY_PRESENT != 0 {
		cb.key1, _ = core.ReadUInt32LE(r)
		cb.key2, _ = core.ReadUInt32LE(r)
	}

	if cb.flags&CBR2_HEIGHT_SAME_AS_WIDTH != 0 {
		cb.bitmapWidth, _ = core.ReadUInt8(r)
		cb.bitmapHeight = cb.bitmapWidth
	} else {
		cb.bitmapWidth, _ = core.ReadUInt8(r)
		cb.bitmapHeight, _ = core.ReadUInt8(r)
	}

	bitmapLength, _ := core.ReadUint16LE(r)
	cacheIndex, _ := core.ReadUInt8(r)

	if cb.flags&CBR2_DO_NOT_CACHE != 0 {
		cb.cacheIndex = 0x7FFF
	} else {
		cb.cacheIndex = uint32(cacheIndex)
	}

	if compressed {
		if cb.flags&CBR2_NO_BITMAP_COMPRESSION_HDR == 0 {
			cb.cbCompFirstRowSize, _ = core.ReadUint16LE(r)
			cb.cbCompMainBodySize, _ = core.ReadUint16LE(r)
			cb.cbScanWidth, _ = core.ReadUint16LE(r)
			cb.cbUncompressedSize, _ = core.ReadUint16LE(r)
			bitmapLength = cb.cbCompMainBodySize
		}
	}

	cb.bitmapDataStream, _ = core.ReadBytes(int(bitmapLength), r)
	cb.bitmapLength = bitmapLength
	cb.compressed = compressed

}

type CacheBitmapV3Order struct {
	cacheId    uint32
	bpp        uint32
	flags      uint32
	cacheIndex uint16
	key1       uint32
	key2       uint32
	bitmapData BitmapDataEx
}
type BitmapDataEx struct {
	bpp     uint8
	codecID uint8
	width   uint16
	height  uint16
	length  uint32
	data    []byte
}

func (s *Secondary) updateCacheBitmapV3Order(r io.Reader, flags uint16) {
	var cb CacheBitmapV3Order

	cb.cacheId = uint32(flags) & 0x00000003
	cb.flags = (uint32(flags) & 0x0000FF80) >> 7
	bitsPerPixelId := (uint32(flags) & 0x00000078) >> 3
	cb.bpp = getCbV2Bpp(bitsPerPixelId)

	cacheIndex, _ := core.ReadUint16LE(r)
	cb.cacheIndex = cacheIndex
	cb.key1, _ = core.ReadUInt32LE(r)
	cb.key2, _ = core.ReadUInt32LE(r)

	bitmapData := &cb.bitmapData
	bitmapData.bpp, _ = core.ReadUInt8(r)
	core.ReadUInt8(r)
	core.ReadUInt8(r)
	bitmapData.codecID, _ = core.ReadUInt8(r)
	bitmapData.width, _ = core.ReadUint16LE(r)
	bitmapData.height, _ = core.ReadUint16LE(r)
	new_len, _ := core.ReadUInt32LE(r)

	bitmapData.data, _ = core.ReadBytes(int(new_len), r)
	bitmapData.length = new_len

}

type CacheColorTableOrder struct {
	cacheIndex   uint8
	numberColors uint16
	colorTable   [256 * 4]uint8
}

func (s *Secondary) updateCacheColorTableOrder(r io.Reader, flags uint16) {
	var cb CacheColorTableOrder
	cb.cacheIndex, _ = core.ReadUInt8(r)
	cb.numberColors, _ = core.ReadUint16LE(r)

	if cb.numberColors != 256 {
		/* This field MUST be set to 256 */
		return
	}

	for i := 0; i < int(cb.numberColors)*4; i++ {
		cb.colorTable[i], cb.colorTable[i+1], cb.colorTable[i+2], cb.colorTable[i+3] = updateReadColorRef(r)
	}
}
func updateReadColorRef(r0 io.Reader) (uint8, uint8, uint8, uint8) {
	b, _ := core.ReadUInt8(r0)
	g, _ := core.ReadUInt8(r0)
	r, _ := core.ReadUInt8(r0)
	core.ReadUInt8(r0)

	return b, g, r, 255
}

type CacheGlyphOrder struct {
	cacheId uint8
	nglyphs uint8
	glyphs  []CacheGlyph
}
type CacheGlyph struct {
	character uint16
	offset    uint16
	baseline  uint16
	width     uint16
	height    uint16
	datasize  int
	data      []uint8
}

func (s *Secondary) updateCacheGlyphOrder(r io.Reader, flags uint16) {
	var cb CacheGlyphOrder

	cb.cacheId, _ = core.ReadUInt8(r)
	cb.nglyphs, _ = core.ReadUInt8(r)
	cb.glyphs = make([]CacheGlyph, 0, cb.nglyphs)

	for i := 0; i < int(cb.nglyphs); i++ {
		var c CacheGlyph
		c.character, _ = core.ReadUint16LE(r)
		c.offset, _ = core.ReadUint16LE(r)
		c.baseline, _ = core.ReadUint16LE(r)
		c.width, _ = core.ReadUint16LE(r)
		c.height, _ = core.ReadUint16LE(r)

		c.datasize = int(c.height*((c.width+7)/8)+3) & ^3
		c.data, _ = core.ReadBytes(c.datasize, r)

		cb.glyphs = append(cb.glyphs, c)
	}
}

type CacheBrushOrder struct {
	index  uint8
	bpp    uint8
	cx     uint8
	cy     uint8
	style  uint8
	length uint8
	data   []uint8
}

func (s *Secondary) updateCacheBrushOrder(r io.Reader, flags uint16) {
	var cb CacheBrushOrder
	cb.index, _ = core.ReadUInt8(r)
	cb.bpp, _ = core.ReadUInt8(r)
	cb.cx, _ = core.ReadUInt8(r)
	cb.cy, _ = core.ReadUInt8(r)
	cb.style, _ = core.ReadUInt8(r)
	cb.length, _ = core.ReadUInt8(r)
	if cb.cx == 8 && cb.cy == 8 {
		if cb.bpp == 1 {
			for i := 7; i >= 0; i-- {
				cb.data[i], _ = core.ReadUInt8(r)
			}
		} else {
			bpp := int(cb.bpp) - 2
			if int(cb.length) == 16+4*bpp {
				/* compressed brush */
				data, _ := core.ReadBytes(int(cb.length), r)
				cb.data = update_decompress_brush(data, bpp)
			} else {
				/* uncompressed brush */
				scanline := 8 * 8 * bpp
				cb.data, _ = core.ReadBytes(scanline, r)
			}
		}
	}
}
func update_decompress_brush(in []uint8, bpp int) []uint8 {
	var pal_index, in_index, shift int

	pal := in[16:]
	out := make([]uint8, 8*8*bpp)
	/* read it bottom up */
	for y := 7; y >= 0; y-- {
		/* 2 bytes per row */
		x := 0
		for do2 := 0; do2 < 2; do2++ {
			/* 4 pixels per byte */
			shift = 6
			for shift >= 0 {
				pal_index = int((in[in_index] >> shift) & 3)
				/* size of palette entries depends on bpp */
				for i := 0; i < bpp; i++ {
					out[(y*8+x)*bpp+i] = pal[pal_index*bpp+i]
				}
				x++
				shift -= 2
			}
			in_index++
		}
	}

	return out
}

/*Primary*/
type Bounds struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}
type OrderInfo struct {
	controlFlags     uint32
	orderType        uint32
	fieldFlags       uint32
	boundsFlags      uint32
	bounds           Bounds
	deltaCoordinates bool
}
