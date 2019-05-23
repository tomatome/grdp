package binary

import "io"

type Writable interface {
	Write(w io.Writer) error
}

type LengthCalculator struct {
	len int
}

func (calc *LengthCalculator) Write(p []byte) (n int, err error) {
	l := len(p)
	calc.len += l
	return l, nil
}

func CalcDataLength(obj Writable) int {
	calc := &LengthCalculator{0}
	obj.Write(calc)
	return calc.len
}
