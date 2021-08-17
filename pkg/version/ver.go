package version

import (
	"fmt"
	"runtime"
)

var Binary string

func String(app string) string {
	return fmt.Sprintf("%s v%s (built w/%s)", app, Binary, runtime.Version())
}
