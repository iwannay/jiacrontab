package libs

import (
	"bufio"

	"bytes"
	"encoding/gob"

	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"

	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

func ReplaceEmpty(str, replaceStr string) string {
	if strings.TrimSpace(str) == "" {
		return replaceStr
	}
	return str
}

func RandNum() int64 {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Int63()
}

func MRecover() {
	if err := recover(); err != nil {
		log.Printf("panic:%s\n%s", err, debug.Stack())
	}
}

func SystemInfo(startTime time.Time) map[string]interface{} {
	var afterLastGC string
	goNum := runtime.NumGoroutine()
	runtime.LockOSThread()
	cpuNum := runtime.NumCPU()
	mstat := &runtime.MemStats{}
	runtime.ReadMemStats(mstat)
	costTime := int(time.Since(startTime).Seconds())
	mb := uint64(1024 * 1024)

	if mstat.LastGC != 0 {
		afterLastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(mstat.LastGC))/1000/1000/1000)
	} else {
		afterLastGC = "0"
	}

	return map[string]interface{}{
		"服务运行时间":    fmt.Sprintf("%d天%d小时%d分%d秒", costTime/(3600*24), costTime%(3600*24)/3600, costTime%3600/60, costTime%(60)),
		"goroute数量": goNum,
		"cpu核心数":    cpuNum,
		"当前内存使用量":   fmt.Sprintf("%.2f MB", float64(mstat.Alloc)/float64(mb)),
		"所有被分配的内存":  fmt.Sprintf("%.2f MB", float64(mstat.TotalAlloc)/float64(mb)),
		"指针查找次数":    mstat.Lookups,
		"内存分配次数":    mstat.Mallocs,
		"内存释放次数":    mstat.Frees,
		"距离上次GC时间":  afterLastGC,
		"下次GC内存回收量": fmt.Sprintf("%.3fMB", float64(mstat.NextGC)/float64(mb)),
		"GC暂停时间总量":  fmt.Sprintf("%.3fs", float64(mstat.PauseTotalNs)/1000/1000/1000),
		"上次GC暂停时间":  fmt.Sprintf("%.3fs", float64(mstat.PauseNs[(mstat.NumGC+255)%256])/1000/1000/1000),
	}
}

func RedirectBack(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, r.Header.Get("Referer"), http.StatusFound)
}

func TryOpen(path string, flag int) (*os.File, error) {
	fabs, err := filepath.Abs(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	f, err := os.OpenFile(fabs, flag, 0644)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(fabs), 0744)
		if err != nil {
			return nil, err
		}
		return os.OpenFile(fabs, flag, 0644)
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

func SortedMap(i map[string]interface{}) {}

func DialHTTP(network, address, path string) (*rpc.Client, error) {
	var err error
	conn, err := net.DialTimeout(network, address, 3*time.Second)
	if err != nil {
		return nil, err
	}
	// err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	// if err != nil {
	// 	return nil, err
	// }
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	connected := "200 Connected to Go RPC"

	if err == nil && resp.Status == connected {
		return rpc.NewClient(conn), nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}
}

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

// DeepCopy2 for map slice
func DeepCopy2(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy2(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy2(v)
		}

		return newSlice
	}

	return value
}

// DeepFind recursive query in the map
func DeepFind(in map[string]interface{}, keyStr string) interface{} {
	if strings.Contains(keyStr, ".") {
		kSlic := strings.Split(keyStr, ".")
		if v, ok := in[kSlic[0]]; ok {
			if v, ok := v.(map[string]interface{}); ok && len(kSlic) >= 2 {
				return DeepFind(v, strings.Join(kSlic[1:], "."))
			}

		}
		return nil
	} else {
		return in[keyStr]
	}

	return nil
}
