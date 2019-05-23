package protocol

import "github.com/chuckpreslar/emission"

type Transport interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error

	On(event, listener interface{}) *emission.Emitter
	Once(event, listener interface{}) *emission.Emitter
	Emit(event interface{}, arguments ...interface{}) *emission.Emitter
}
