package version

import (
	"fmt"
	"runtime"
)

const Binary = "2.0.0-alpha726"

func String(app string) string {
	return fmt.Sprintf("%s v%s (built w/%s)", app, Binary, runtime.Version())
}
