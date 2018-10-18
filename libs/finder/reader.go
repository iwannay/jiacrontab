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

	t.curr = t.curr - int64(len(b))
	if t.curr < 0 {
		t.curr = 0
	}

	n, err = t.f.ReadAt(b, t.curr)
	if err != nil && err != io.EOF {
		return n, err
	}

	invertSplit(b, n)

	if t.curr == 0 {
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

func invertSplit(b []byte, n int) {
	invert(b[0:n])
	var k, i int

	for k, i = 0, 0; k < n; k++ {
		if b[k] == '\n' && n > k+1 && b[k+1] == '\r' {
			b[k], b[k+1] = '\r', '\n'
			k++
			if i != k {
				invert(b[i:k])
			}

			i = k + 1
			continue
		} else if b[k] == '\n' {
			if i != k {
				invert(b[i:k])
			}
			i = k + 1
		}
	}

	if i != k {
		invert(b[i:k])
	}

}
