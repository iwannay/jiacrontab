// +build !windows

package log

var levelMap = map[int]string{
	LevelDebug: "[DEBUG] ",
	LevelInfo:  "[INFO] ",
	LevelWarn:  "[WARN] ", // 黄
	LevelError: "[ERROR] ",
	LevelFatal: "[FATAL] ",
}
