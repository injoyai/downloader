package spider

import (
	_ "embed"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
)

func Install(bs []byte) error {

	//下载文件
	if oss.Exists("./browser/chrome/chrome.exe") {
		return nil
	}

	if err := oss.New("./browser/chrome.zip", bs); err != nil {
		return err
	}
	defer oss.Remove("./browser/chrome.zip")
	defer oss.Remove("./browser/hrome")

	if err := zip.Decode("./browser/chrome.zip", "./browser/"); err != nil {
		return err
	}

	return nil
}
