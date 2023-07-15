package spider

import (
	"errors"
	"fmt"
	"github.com/injoyai/downloader/tool"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/win"
	"github.com/injoyai/goutil/str"
	"io/ioutil"
	"os"
	"strings"
)

func Install() error {
	if oss.Exists("./chromedriver.exe") {
		return nil
	}
	appPath := win.AppPath("chrome.exe")
	if len(appPath) == 0 {
		return errors.New("需要安装chrome")
	}

	version := ""
	path := str.CropLast(appPath[0], "\\")
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, v := range dirs {
		if v.IsDir() {
			list := strings.Split(v.Name(), ".")
			if len(list) > 3 {
				version = v.Name()
				break
			}
		}
	}

	if list := strings.Split(version, "."); len(list) >= 2 {
		url := getVersion(list[0])
		bs, err := http.GetBytes(url)
		if err != nil {
			return err
		}
		zipPath := "./chromedriver.zip"
		if err := oss.New(zipPath, bs); err != nil {
			return err
		}
		if err := tool.DecodeZip(zipPath, "./"); err != nil {
			return err
		}
		if err := os.Remove(zipPath); err != nil {
			return err
		}
	}

	return nil
}

func getVersion(version string) string {
	//http://chromedriver.storage.googleapis.com/index.html
	url := "http://chromedriver.storage.googleapis.com/%s/chromedriver_win32.zip"
	v := "96.0.4664.45"
	switch version {
	case "70":
		v = "70.0.3538.97"
	case "71":
		v = "71.0.3578.80"
	case "72":
		v = "72.0.3626.7"
	case "73":
		v = "73.0.3683.68"
	case "74":
		v = "74.0.3729.6"
	case "75":
		v = "75.0.3770.90"
	case "76":
		v = "76.0.3809.68"
	case "77":
		v = "77.0.3865.40"
	case "78":
		v = "78.0.3904.70"
	case "79":
		v = "79.0.3945.36"
	case "80":
		v = "80.0.3987.16"
	case "81":
		v = "81.0.4044.69"
	case "83":
		v = "83.0.4103.39"
	case "84":
		v = "84.0.4147.30"
	case "85":
		v = "85.0.4183.87"
	case "86":
		v = "86.0.4240.22"
	case "87":
		v = "87.0.4280.88"
	case "88":
		v = "88.0.4324.96"
	case "89":
		v = "89.0.4389.23"
	case "90":
		v = "90.0.4430.24"
	case "91":
		v = "91.0.4472.19"
	case "92":
		v = "92.0.4515.43"
	case "93":
		v = "93.0.4577.63"
	case "94":
		v = "94.0.4606.61"
	case "95":
		v = "95.0.4638.69"
	case "96":
		v = "96.0.4664.45"
	case "97":
		v = "97.0.4692.71"
	case "98":
		v = "98.0.4758.80"
	case "99":
		v = "99.0.4844.51"
	case "100":
		v = "100.0.4896.60"
	case "101":
		v = "101.0.4951.41"
	case "102":
		v = "102.0.5005.61"
	case "103":
		v = "103.0.5060.53"
	case "104":
		v = "104.0.5112.79"
	case "105":
		v = "105.0.5195.52"
	case "106":
		v = "106.0.5249.21"
	case "107":
		v = "107.0.5304.62"
	case "108":
		v = "108.0.5359.71"
	case "109":
		v = "109.0.5414.74"
	case "110":
		v = "110.0.5481.77"
	case "112":
		v = "112.0.5615.49"
	case "113":
		v = "113.0.5672.63"
	case "114":
		v = "114.0.5735.16"
	}
	return fmt.Sprintf(url, v)
}
