package lib

func StrConcat(str ...string) string {
	bs := make([]byte, 40)
	bl := 0
	for _, s := range str {
		bl += copy(bs[bl:], s)
	}
	return string(bs[:bl])
}
