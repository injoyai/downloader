package spider

import (
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"path/filepath"
)

func Install(browserDir string, chromeZip []byte) error {

	chromeZipPath := filepath.Join(browserDir, "chrome.zip")
	chromeExePath := filepath.Join(browserDir, "chrome/chrome.exe")

	if oss.Exists(chromeExePath) {
		return nil
	}

	if err := oss.New(chromeZipPath, chromeZip); err != nil {
		return err
	}
	defer oss.Remove(chromeZipPath)

	if err := zip.Decode(chromeZipPath, browserDir); err != nil {
		return err
	}

	return nil
}
