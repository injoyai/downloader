package tool

import "strings"

func CropFirst(str, crop string, retain ...bool) string {
	n := strings.Index(str, crop)
	if n >= 0 {
		if len(retain) > 0 && !retain[0] {
			return str[n+len(crop):]
		}
		return str[n:]
	}
	return str
}

// CropLast 裁剪,剪短
// 例: "0123456789", "2" >>> "012"
// @str,被裁剪的字符串
// @crop,裁剪的字符串
// @retain,是否保留裁剪字符串,默认保留 "0123456789", "2" >>> "012" "0123456789", "2" >>> "01"
func CropLast(str, crop string, retain ...bool) string {
	n := strings.LastIndex(str, crop)
	if n+len(crop) <= len(str) && n >= 0 {
		if len(retain) > 0 && !retain[0] {
			return str[:n]
		}
		return str[:n+len(crop)]
	}
	return str
}
