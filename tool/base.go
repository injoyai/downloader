package tool

import (
	"archive/zip"
)

// Decode 解压zip
func Decode(zipPath, filePath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, k := range r.Reader.File {
		var err error
		if k.FileInfo().IsDir() {
			NewFile(filePath + k.Name[1:])
		} else {
			r, err := k.Open()
			if err == nil {
				err = NewFile(filePath+"/"+k.Name, r)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
