package utils

func IntMin(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func RuneEqua(a, b []rune) bool {

	l := len(a)
	if l != len(b) {
		return false
	}

	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true

}
