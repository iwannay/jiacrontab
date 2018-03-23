package base

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iwannay/jiaweb/utils"
)

const (
	minuteTimeLayout        = "200601021504"
	dateTimeLayout          = "2006-01-02 15:04:05"
	defaultReserveMinuts    = 60
	defaultCheckTimeMinutes = 10
)

var GlobalState *State

type (
	State struct {
		ServerStartTime         time.Time
		EnableDetailRequestData bool
		TotalRequestCount       uint64

		IntervalRequestData *Store
		DetailRequstURLData *Store
		TotalErrorCount     uint64
		IntervalErrorData   *Store
		DetailErrorPageData *Store
		DetailErrorData     *Store
		DetailHTTPCodeData  *Store

		dataChanRequest  chan *RequestInfo
		dataChanError    chan *ErrorInfo
		dataChanHttpCode chan *HttpCodeInfo

		infoPool *pool
	}

	pool struct {
		requestInfo  sync.Pool
		errorInfo    sync.Pool
		httpCodeInfo sync.Pool
	}

	RequestInfo struct {
		URL  string
		Code int
		Num  uint64
	}

	ErrorInfo struct {
		URL    string
		ErrMsg string
		Num    uint64
	}

	HttpCodeInfo struct {
		URL  string
		Code int
		Num  uint64
	}
)

func (s *State) QueryIntervalRequstData(key string) uint64 {
	val, _ := s.IntervalRequestData.GetUint64(key)
	return val
}

func (s *State) QueryIntervalErrorData(key string) uint64 {
	val, _ := s.IntervalErrorData.GetUint64(key)
	return val
}

func (s *State) AddRequestCount(page string, code int, num uint64) uint64 {
	if strings.Index("page", "jiaweb") != 0 {
		atomic.AddUint64(&s.TotalRequestCount, num)
		s.addRequestData(page, code, num)
		s.AddHTTPCodeData(page, code, num)
	}
	return atomic.LoadUint64(&s.TotalRequestCount)
}

func (s *State) AddErrorCount(page string, err error, num uint64) uint64 {
	atomic.AddUint64(&s.TotalErrorCount, num)
	s.addErrorData(page, err, num)
	return atomic.LoadUint64(&s.TotalErrorCount)
}

func (s *State) addRequestData(page string, code int, num uint64) {
	info := s.infoPool.requestInfo.Get().(*RequestInfo)
	info.URL = page
	info.Code = code
	info.Num = num
	s.dataChanRequest <- info
}

func (s *State) addErrorData(page string, err error, num uint64) {
	info := s.infoPool.errorInfo.Get().(*ErrorInfo)
	info.URL = page
	info.ErrMsg = err.Error()
	info.Num = num
	s.dataChanError <- info
}

func (s *State) AddHTTPCodeData(page string, code int, num uint64) {
	info := s.infoPool.httpCodeInfo.Get().(*HttpCodeInfo)
	info.URL = page
	info.Code = code
	info.Num = num
	s.dataChanHttpCode <- info
}

func (s *State) handleInfo() {
	for {
		select {
		case info := <-s.dataChanRequest:
			{
				if s.EnableDetailRequestData {
					if info.Code != http.StatusNotFound {
						key := strings.ToLower(info.URL)
						val, _ := s.DetailRequstURLData.GetUint64(key)
						s.DetailRequstURLData.Store(key, val+info.Num)
					}
				}

				key := time.Now().Format(minuteTimeLayout)
				val, _ := s.IntervalRequestData.GetUint64(key)
				s.IntervalRequestData.Store(key, val+info.Num)

				s.infoPool.requestInfo.Put(info)
			}
		case info := <-s.dataChanError:
			{
				key := strings.ToLower(info.URL)
				val, _ := s.DetailErrorPageData.GetUint64(key)
				s.DetailErrorPageData.Store(key, val+info.Num)

				key = info.ErrMsg

				val, _ = s.DetailErrorData.GetUint64(key)

				s.DetailErrorData.Store(key, val+info.Num)

				key = time.Now().Format(minuteTimeLayout)
				val, _ = s.IntervalErrorData.GetUint64(key)
				s.IntervalErrorData.Store(key, val+info.Num)

				s.infoPool.errorInfo.Put(info)

			}

		case info := <-s.dataChanHttpCode:
			{
				key := strconv.Itoa(info.Code)
				val, _ := s.DetailHTTPCodeData.GetUint64(key)
				s.DetailHTTPCodeData.Store(key, val+info.Num)

				s.infoPool.httpCodeInfo.Put(info)
			}
		}
	}
}

func (s *State) ShowData(dataType string) string {

	if dataType == "json" {
		var dataMap = make(map[string]interface{})
		dataMap["ServerStartTime"] = s.ServerStartTime.Format(dateTimeLayout)
		dataMap["TotalRequestCount"] = strconv.FormatUint(s.TotalRequestCount, 10)
		dataMap["TotalErrorCount"] = strconv.FormatUint(s.TotalErrorCount, 10)
		dataMap["IntervalRequestData"] = s.IntervalRequestData.All()
		dataMap["DetailRequestUrlData"] = s.DetailRequstURLData.All()
		dataMap["IntervalErrorData"] = s.IntervalErrorData.All()
		dataMap["DetailErrorPageData"] = s.DetailErrorPageData.All()
		dataMap["DetailErrorData"] = s.DetailErrorData.All()
		dataMap["DetailHttpCodeData"] = s.DetailHTTPCodeData.All()
		return utils.GetJsonString(dataMap)
	}

	data := "<html><body><div>"
	data += "ServerStartTime : " + s.ServerStartTime.Format(dateTimeLayout)
	data += "<br>"
	data += "TotalRequestCount : " + strconv.FormatUint(s.TotalRequestCount, 10)
	data += "<br>"
	data += "TotalErrorCount : " + strconv.FormatUint(s.TotalErrorCount, 10)
	data += "<br>"
	data += "IntervalRequestData : " + utils.GetJsonString(s.IntervalRequestData.All())
	data += "<br>"

	data += "DetailRequestUrlData : " + utils.GetJsonString(s.DetailRequstURLData.All())
	data += "<br>"

	data += "IntervalErrorData : " + utils.GetJsonString(s.IntervalErrorData.All())

	data += "<br>"

	data += "DetailErrorPageData : " + utils.GetJsonString(s.DetailErrorPageData.All())

	data += "<br>"

	data += "DetailErrorData : " + utils.GetJsonString(s.DetailErrorData.All())

	data += "<br>"

	data += "DetailHttpCodeData : " + utils.GetJsonString(s.DetailHTTPCodeData.All())

	data += "</div></body></html>"
	return data

}

func (s *State) gc() {
	var needRemoveKey []string
	now, _ := time.Parse(minuteTimeLayout, time.Now().Format(minuteTimeLayout))

	if s.IntervalRequestData.Len() > defaultReserveMinuts {
		s.IntervalRequestData.Range(func(key, val interface{}) bool {
			keyString := key.(string)
			if t, err := time.Parse(minuteTimeLayout, keyString); err != nil {
				needRemoveKey = append(needRemoveKey, keyString)
			} else {
				if now.Sub(t) > (defaultReserveMinuts * time.Minute) {
					needRemoveKey = append(needRemoveKey, keyString)
				}
			}
			return true
		})
	}

	for _, v := range needRemoveKey {
		s.IntervalRequestData.Delete(v)
	}

	needRemoveKey = []string{}
	if s.IntervalErrorData.Len() > defaultReserveMinuts {
		s.IntervalErrorData.Range(func(key, val interface{}) bool {
			keyString := key.(string)
			if t, err := time.Parse(minuteTimeLayout, keyString); err != nil {
				needRemoveKey = append(needRemoveKey, keyString)
			} else {
				if now.Sub(t) > defaultReserveMinuts*time.Minute {
					needRemoveKey = append(needRemoveKey, keyString)
				}
			}
			return true
		})

	}

	for _, v := range needRemoveKey {
		s.IntervalErrorData.Delete(v)
	}

	time.AfterFunc(time.Duration(defaultCheckTimeMinutes)*time.Minute, s.gc)

}

func init() {
	GlobalState = &State{
		ServerStartTime:     time.Now(),
		IntervalRequestData: NewStore(),
		DetailRequstURLData: NewStore(),
		IntervalErrorData:   NewStore(),
		DetailErrorPageData: NewStore(),
		DetailErrorData:     NewStore(),
		DetailHTTPCodeData:  NewStore(),
		dataChanRequest:     make(chan *RequestInfo, 1000),
		dataChanError:       make(chan *ErrorInfo, 1000),
		dataChanHttpCode:    make(chan *HttpCodeInfo, 1000),
		infoPool: &pool{
			requestInfo: sync.Pool{
				New: func() interface{} {
					return &RequestInfo{}
				},
			},
			errorInfo: sync.Pool{
				New: func() interface{} {
					return &ErrorInfo{}
				},
			},
			httpCodeInfo: sync.Pool{
				New: func() interface{} {
					return &HttpCodeInfo{}
				},
			},
		},
	}

	go GlobalState.handleInfo()
	go time.AfterFunc(time.Duration(defaultCheckTimeMinutes)*time.Minute, GlobalState.gc)
}
