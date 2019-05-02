package finder

import (
	"io"
	"os"
)

type TailReader struct {
	f     *os.File
	curr  int64
	isEOF bool
}

func (t *TailReader) Read(b []byte) (n int, err error) {
	if t.isEOF {
		return 0, io.EOF
	}

	off := t.curr - int64(len(b))
	if off < 0 {
		off = 0
		n, err = t.f.ReadAt(b[0:t.curr], off)
	} else {
		t.curr = off
		n, err = t.f.ReadAt(b, off)
	}

	if err != nil && err != io.EOF {
		return n, err
	}

	invert(b[0:n])

	if off == 0 {
		t.isEOF = true
	}

	return
}

func NewTailReader(f *os.File, offset int64) io.Reader {
	return &TailReader{
		f:    f,
		curr: offset,
	}

}

func invert(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
