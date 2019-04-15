package util

import (
	"jiacrontab/pkg/file"
	"reflect"
	"strconv"

	"github.com/gofrs/uuid"

	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/iwannay/log"

	"os"
	"path/filepath"
	"runtime"
	"time"
)

func RandIntn(end int) int {
	return rand.Intn(end)
}

func CurrentTime(t int64) string {
	if t == 0 {
		return "0"
	}
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

func SystemInfo(startTime time.Time) map[string]interface{} {
	var afterLastGC string
	goNum := runtime.NumGoroutine()
	cpuNum := runtime.NumCPU()
	mstat := &runtime.MemStats{}
	runtime.ReadMemStats(mstat)
	costTime := int(time.Since(startTime).Seconds())
	mb := 1024 * 1024

	if mstat.LastGC != 0 {
		afterLastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(mstat.LastGC))/1000/1000/1000)
	} else {
		afterLastGC = "0"
	}

	return map[string]interface{}{
		"服务运行时间":      fmt.Sprintf("%d天%d小时%d分%d秒", costTime/(3600*24), costTime%(3600*24)/3600, costTime%3600/60, costTime%(60)),
		"goroutine数量": goNum,
		"cpu核心数":      cpuNum,

		"当前内存使用量":  file.FileSize(int64(mstat.Alloc)),
		"所有被分配的内存": file.FileSize(int64(mstat.TotalAlloc)),
		"内存占用量":    file.FileSize(int64(mstat.Sys)),
		"指针查找次数":   mstat.Lookups,
		"内存分配次数":   mstat.Mallocs,
		"内存释放次数":   mstat.Frees,
		"距离上次GC时间": afterLastGC,

		// "当前 Heap 内存使用量": file.FileSize(int64(mstat.HeapAlloc)),
		// "Heap 内存占用量":    file.FileSize(int64(mstat.HeapSys)),
		// "Heap 内存空闲量":    file.FileSize(int64(mstat.HeapIdle)),
		// "正在使用的 Heap 内存": file.FileSize(int64(mstat.HeapInuse)),
		// "被释放的 Heap 内存":  file.FileSize(int64(mstat.HeapReleased)),
		// "Heap 对象数量":     mstat.HeapObjects,

		"下次GC内存回收量": fmt.Sprintf("%.3fMB", float64(mstat.NextGC)/float64(mb)),
		"GC暂停时间总量":  fmt.Sprintf("%.3fs", float64(mstat.PauseTotalNs)/1000/1000/1000),
		"上次GC暂停时间":  fmt.Sprintf("%.3fs", float64(mstat.PauseNs[(mstat.NumGC+255)%256])/1000/1000/1000),
	}
}

func TryOpen(path string, flag int) (*os.File, error) {
	fabs, err := filepath.Abs(path)
	if err != nil {
		log.Errorf("TryOpen:", err)
		return nil, err
	}

	f, err := os.OpenFile(fabs, flag, os.ModePerm)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(fabs), os.ModePerm)
		if err != nil {
			return nil, err
		}
		return os.OpenFile(fabs, flag, os.ModePerm)
	}
	return f, err
}

func CatFile(filepath string, limit int64, content *string) (isPath bool, err error) {
	f, err := os.Open(filepath)

	if err != nil {
		return false, err
	}
	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	if fi.Size() > limit {
		*content = filepath
		return true, nil
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return false, err
	}
	*content = string(data)
	return false, nil
}

func ParseInt(i string) int {
	v, _ := strconv.Atoi(i)
	return v
}

func ParseInt64(i string) int64 {
	v, _ := strconv.Atoi(i)
	return int64(v)
}

func InArray(val interface{}, arr interface{}) bool {
	t := reflect.TypeOf(arr)
	v := reflect.ValueOf(arr)

	if t.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() == val {
				return true
			}
		}
	}

	return false
}

func UUID() string {

	uu, err := uuid.NewGen().NewV1()

	if err != nil {
		log.Error(err)
		return fmt.Sprint(time.Now().UnixNano())
	}

	return uu.String()
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error("GetHostname:", err)
	}
	return hostname
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
