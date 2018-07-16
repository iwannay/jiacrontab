package finder

import (
	"bufio"
	"errors"
	"jiacrontab/libs/file"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
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
	matchDataQueue    DataQueue
	curr              uint64
	regexp            *regexp.Regexp
	seekCurr, seekEnd int
	maxRows           uint64
	errors            []error
	patternAll        bool
	filter            func(os.FileInfo) bool
	group             sync.WaitGroup
}

func NewFinder(maxRows uint64, filter func(os.FileInfo) bool) *Finder {
	return &Finder{
		maxRows: maxRows,
		filter:  filter,
	}
}

func (fd *Finder) Count() uint64 {
	return fd.curr
}

func (fd *Finder) find(fpath string, modifyTime time.Time) error {

	var matchData []byte

	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		if atomic.LoadUint64(&fd.curr) >= fd.maxRows {
			break
		}

		bts, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		if fd.patternAll || fd.regexp.Match(bts) {
			if fd.curr >= uint64(fd.seekCurr) && fd.curr <= uint64(fd.seekEnd) {
				matchData = append(matchData, bts...)
				matchData = append(matchData, []byte("\n")...)
			}
			atomic.AddUint64(&fd.curr, 1)
		}

	}
	if len(matchData) > 0 {
		fd.matchDataQueue = append(fd.matchDataQueue, matchDataChunk{
			modifyTime: modifyTime,
			matchData:  matchData,
		})
	}
	return nil
}

func (fd *Finder) walkFunc(fpath string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		if fd.filter != nil && fd.filter(info) {
			fd.group.Add(1)
			go func() {
				defer fd.group.Done()
				err := fd.find(fpath, info.ModTime())
				if err != nil {
					fd.errors = append(fd.errors, err)
				}
			}()
		}

	}

	return nil

}

func (fd *Finder) Search(root string, expr string, data *[]byte, page, pagesize int) error {
	var err error
	fd.seekCurr = (page - 1) * pagesize
	fd.seekEnd = fd.seekCurr + pagesize
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
	fd.group.Wait()
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
