// +build !windows

package log

var logLevels = []string{
	LevelDebug: "[DEBUG]",
	LevelInfo:  "[INFO]",
	LevelWarn:  "[WARN]", // 黄
	LevelError: "[ERROR]",
	LevelFatal: "[FATAL]",
	LevelPrint: "",
}

func closeColor(content string) string {
	return content
}
