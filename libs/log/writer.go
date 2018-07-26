package log

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type WriterOptions struct {
	Dir            string
	FileNameFormat string
	Size           int64
	Prefix         string
	Suffix         string
}

type Writer struct {
	w      *bufio.Writer
	n      int64
	size   int64
	dir    string
	prefix string
	suffix string
	prev   string

	f *os.File
}

func NewWriter(opt *WriterOptions) *Writer {
	w := &Writer{
		size:   opt.Size,
		dir:    opt.Dir,
		prefix: opt.Prefix,
		suffix: opt.Suffix,
	}

	err := w.Open()
	if err != nil {
		panic(err)
	}
	w.w = bufio.NewWriter(w.f)

	return w
}

func (w *Writer) Index() (int, error) {
	fs, err := ioutil.ReadDir(w.dir)
	if err != nil {
		return 0, err
	}

	var index int
	for _, f := range fs {
		name := f.Name()
		if !strings.HasPrefix(name, w.prefix) || !strings.HasSuffix(name, w.suffix) {
			continue
		}
		pos := strings.TrimSuffix(strings.TrimPrefix(f.Name(), w.prefix), w.suffix)
		i, err := strconv.Atoi(pos)
		if err != nil {
			continue
		}
		if i > index {
			index = i
		}
	}
	return index, nil
}

func (w *Writer) Open() error {
	var (
		err   error
		i     int
		path  string
		now   time.Time
		fInfo os.FileInfo
	)
	err = os.MkdirAll(w.dir, os.ModePerm)
	if err != nil {
		return err
	}

	if i, err = w.Index(); err != nil {
		return err
	}

	now = time.Now()

	prefix, suffix := now.Format(w.prefix), now.Format(w.suffix)
	w.prev = prefix

	// if i == 0 {
	// 	path = filepath.Join(w.dir, prefix+suffix)
	// } else {
	path = filepath.Join(w.dir, fmt.Sprint(prefix, i, suffix))
	// }

	w.f, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err == nil {
		fInfo, err = w.f.Stat()
		if err == nil {
			w.n = fInfo.Size()
		}
	}
	return err
}

func (w *Writer) Reset() error {
	err := w.w.Flush()
	w.f.Close()

	if err != nil {
		return err
	}

	err = w.Open()
	if err != nil {
		return err
	}
	w.n = 0
	w.w = bufio.NewWriter(w.f)
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	if w.n >= w.size {
		err := w.Reset()
		if err != nil {
			return 0, err
		}
		l, err := w.w.Write(b)
		w.n += int64(l)
		return l, err
	}

	if time.Now().Format(w.prefix) != w.prev {
		err := w.Reset()
		if err != nil {
			return 0, err
		}
		l, err := w.w.Write(b)
		w.n += int64(l)
		return l, err
	}

	l, err := w.w.Write(b)
	w.n += int64(l)

	return l, err
}
