package file

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iwannay/log"
)

func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
		return ""
	}
	return filepath.Clean(strings.Replace(dir, "\\", "/", -1))

}

func IsTextFile(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return strings.Contains(http.DetectContentType(data), "text/")
}
func IsImageFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "image/")
}
func IsPDFFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "application/pdf")
}
func IsVideoFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "video/")
}

const (
	Byte  = 1
	KByte = Byte * 1024
	MByte = KByte * 1024
	GByte = MByte * 1024
	TByte = GByte * 1024
	PByte = TByte * 1024
	EByte = PByte * 1024
)

var bytesSizeTable = map[string]uint64{
	"b":  Byte,
	"kb": KByte,
	"mb": MByte,
	"gb": GByte,
	"tb": TByte,
	"pb": PByte,
	"eb": EByte,
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}
	return fmt.Sprintf(f+" %s", val, suffix)
}
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}

func CreateFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

func DirSize(dir string) int64 {
	var total int64
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func Remove(dir string, t time.Time) (total int64, size int64, err error) {
	err = filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if info == nil {
			return errors.New(fpath + " not exists")
		}
		if !info.IsDir() {
			if info.ModTime().Before(t) {
				total++
				size += info.Size()
				err = os.Remove(fpath)
			}
		} else {
			// 删除空目录
			err = os.Remove(fpath)
			if err == nil {
				total++
				log.Println("delete ", fpath)
			}
			err = nil
		}

		return err
	})
	return
}
