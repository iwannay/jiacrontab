package finder

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
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

// var (
// 	matchDataQueue DataQueue
// 	group          sync.WaitGroup
// )

type Finder struct {
	matchDataQueue DataQueue
	expr           string
	filterExt      string
	group          sync.WaitGroup
}

func NewFinder(expr, filterExt string) *Finder {
	return &Finder{
		expr:      expr,
		filterExt: filterExt,
	}
}

func (fd *Finder) find(fpath string, modifyTime time.Time) error {
	fd.group.Add(1)
	defer fd.group.Done()
	var matchData []byte
	re, err := regexp.Compile(fd.expr)
	if err != nil {
		return err
	}

	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	for {
		bts, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		if re.Match(bts) {
			matchData = append(matchData, re.ReplaceAll(bts, []byte(`<span style="color:red">$0</span>`))...)
			matchData = append(matchData, []byte("\n")...)
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
		if filepath.Ext(fpath) == ".go" {
			go fd.find(fpath, info.ModTime())
		}

	}

	return nil

}

func (fd *Finder) Search(root string, data *[]byte) {
	filepath.Walk("/home/john/goproject/src/jiacrontab", fd.walkFunc)

	fd.group.Wait()

	sort.Stable(fd.matchDataQueue)

	for _, v := range fd.matchDataQueue {
		*data = append(*data, v.matchData...)
	}

}
