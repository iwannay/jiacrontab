// +build !windows

package log

import "fmt"

var logLevels = []string{
	LevelDebug: "\033[36m[DEBUG]\033[0m",
	LevelInfo:  "\033[32m[INFO]\033[0m",
	LevelWarn:  "\033[33m[WARN]\033[0m",
	LevelError: "\033[31m[ERROR]\033[0m",
	LevelFatal: "\033[35m[FATAL]\033[0m",
	LevelPrint: "\033[33m",
}

func closeColor(content string) string {
	return fmt.Sprint(content, "\033[0m")
}
