// +build !windows

package log

var levelMap = map[int]string{
	LevelDebug: "\033[36m [DEBUG];33[0m ",
	LevelInfo:  "\033[32m [INFO];33[0m ",
	LevelWarn:  "\033[33m [WARN];33[0m ", // 黄
	LevelError: "\033[31m [ERROR];33[0m ",
	LevelFatal: "\033[35m [FATAL];33[0m ",
}
