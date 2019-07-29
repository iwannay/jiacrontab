package formbinder

import (
	"github.com/gorilla/schema"
)

var defaultDecoder = schema.NewDecoder()

func init() {
	defaultDecoder.SetAliasTag("form")
}

// IsErrPath reports whether the incoming error is type of unknown field passed,
// which can be ignored when server allows unknown post values to be sent by the client.
func IsErrPath(err error) bool {
	if err == nil {
		return false
	}

	if m, ok := err.(schema.MultiError); ok {
		j := len(m)
		for _, e := range m {
			if _, is := e.(schema.UnknownKeyError); is {
				j--
			}
		}

		return j == 0
	}

	return false
}

// Decode maps "values" to "ptr".
func Decode(values map[string][]string, ptr interface{}) error {
	return defaultDecoder.Decode(ptr, values)
}
