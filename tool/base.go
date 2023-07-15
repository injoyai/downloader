package tool

import (
	"archive/zip"
	"github.com/injoyai/goutil/oss"
)

// DecodeZip 解压zip
func DecodeZip(zipPath, filePath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, k := range r.Reader.File {
		var err error
		if k.FileInfo().IsDir() {
			oss.New(filePath + k.Name[1:])
		} else {
			r, err := k.Open()
			if err == nil {
				err = oss.New(filePath+"/"+k.Name, r)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
