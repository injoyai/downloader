package tool

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// NewFile 新建文件,会覆盖
func NewFile(path string, v ...interface{}) error {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	if len(name) == 0 {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if len(v) == 0 {
		return nil
	}
	data := []byte(nil)
	for _, k := range v {
		bs := []byte(nil)
		switch val := k.(type) {
		case []byte:
			bs = val
		case string:
			bs = []byte(val)
		case io.Reader:
			bs, _ = ioutil.ReadAll(val)
		default:
			bs = []byte(fmt.Sprint(k))
		}
		data = append(data, bs...)
	}
	_, err = f.Write(data)
	return err
}
