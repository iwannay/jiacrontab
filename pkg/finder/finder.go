package finder

import (
	"bufio"
	"bytes"
	"errors"
	"jiacrontab/pkg/file"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type matchDataChunk struct {
	modifyTime time.Time
	matchData  []byte
}

type DataQueue []matchDataChunk

func (d DataQueue) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d DataQueue) Less(i, j int) bool {
	return d[i].modifyTime.Unix() < d[j].modifyTime.Unix()
}
func (d DataQueue) Len() int {
	return len(d)
}

type Finder struct {
	matchDataQueue DataQueue
	curr           int32
	regexp         *regexp.Regexp
	pagesize       int
	errors         []error
	patternAll     bool
	filter         func(os.FileInfo) bool
	isTail         bool
	offset         int64
	fileSize       int64
}

func NewFinder(filter func(os.FileInfo) bool) *Finder {
	return &Finder{
		filter: filter,
	}
}

func (fd *Finder) SetTail(flag bool) {
	fd.isTail = flag
}

func (fd *Finder) Offset() int64 {
	return fd.offset
}

func (fd *Finder) HumanateFileSize() string {
	return file.FileSize(fd.fileSize)
}

func (fd *Finder) FileSize() int64 {
	return fd.fileSize
}

func (fd *Finder) find(fpath string, modifyTime time.Time) error {

	var matchData []byte
	var reader *bufio.Reader

	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fd.fileSize = info.Size()

	if fd.fileSize < fd.offset {
		return errors.New("out of file")
	}

	if fd.isTail {
		if fd.offset == 0 {
			fd.offset = fd.fileSize
		}
		f.Seek(fd.offset, 0)
		reader = bufio.NewReader(NewTailReader(f, fd.offset))
	} else {
		f.Seek(fd.offset, 0)
		reader = bufio.NewReader(f)
	}

	for {

		bts, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		if fd.isTail {
			fd.offset -= int64(len(bts))
		} else {
			fd.offset += int64(len(bts))
		}

		if fd.isTail {
			invert(bts)
		}

		if fd.patternAll || fd.regexp.Match(bts) {
			matchData = append(matchData, bts...)
			fd.curr++
		}

		if fd.curr >= int32(fd.pagesize) {
			break
		}

		if fd.offset == 0 {
			break
		}

	}
	if len(matchData) > 0 {
		fd.matchDataQueue = append(fd.matchDataQueue, matchDataChunk{
			modifyTime: modifyTime,
			matchData:  bytes.TrimLeft(bytes.TrimRight(matchData, "\n"), "\n"),
		})
	}
	return nil
}

func (fd *Finder) walkFunc(fpath string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		if fd.filter != nil && fd.filter(info) {
			err := fd.find(fpath, info.ModTime())
			if err != nil {
				fd.errors = append(fd.errors, err)
			}
		}

	}

	return nil
}

func (fd *Finder) Search(root string, expr string, data *[]byte, offset int64, pagesize int) error {
	var err error
	fd.pagesize = pagesize
	fd.offset = offset

	if expr == "" {
		fd.patternAll = true
	}

	if !file.Exist(root) {
		return errors.New(root + " not exist")
	}

	fd.regexp, err = regexp.Compile(expr)
	if err != nil {
		return err
	}
	filepath.Walk(root, fd.walkFunc)

	sort.Stable(fd.matchDataQueue)
	for _, v := range fd.matchDataQueue {
		*data = append(*data, v.matchData...)
	}
	return nil
}

func (fd *Finder) GetErrors() []error {
	return fd.errors
}

func SearchAndDeleteFileOnDisk(dir string, d time.Duration, size int64) {
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
				if info == nil {
					return errors.New(fpath + "not exists")
				}
				if !info.IsDir() {
					if time.Now().Sub(info.ModTime()) > d {
						os.Remove(fpath)
						return nil
					}

					if info.Size() > size && size != 0 {
						os.Remove(fpath)
						return nil
					}
				}

				if info.IsDir() {
					// 删除空目录
					err := os.Remove(fpath)
					if err == nil {
						log.Println("delete ", fpath)
					}
				}

				return nil
			})
		}
	}
}
