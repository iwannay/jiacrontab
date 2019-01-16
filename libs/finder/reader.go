package finder

import (
	"io"
	"os"
)

type TailReader struct {
	f     *os.File
	curr  int64
	size  int64
	isEOF bool
}

func (t *TailReader) Read(b []byte) (n int, err error) {

	var fi os.FileInfo
	if t.size == 0 {
		fi, err = t.f.Stat()
		if err != nil {
			return
		}
		t.size = fi.Size()
		t.curr = t.size
	}

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

func NewTailReader(f *os.File) io.Reader {
	return &TailReader{
		f: f,
	}

}

func invert(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
