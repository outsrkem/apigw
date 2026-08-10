package uuid

// V4UUID 生成uuid
func V4UUID() string {
	uuidV4, err := NewV4()
	if err != nil {
		panic(err)
	}
	return uuidV4.NoHyphenString()
}
