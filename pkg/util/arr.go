package util

func FilterEmptyEle(in []string) (out []string) {
	for _, v := range in {
		if v != "" {
			out = append(out, v)
		}
	}
	return
}
