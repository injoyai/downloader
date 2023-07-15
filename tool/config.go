package tool

import (
	"encoding/json"
	"github.com/injoyai/goutil/oss"
	"io/ioutil"
	"os/user"
)

var (
	cfgPath = func() string {
		u, _ := user.Current()
		return u.HomeDir + "/AppData/Local/downloader/config.json"
	}()
)

var Cfg = func() *cfg {
	bs, _ := ioutil.ReadFile(cfgPath)
	data := new(cfg)
	json.Unmarshal(bs, data)
	return data
}()

type cfg struct {
	Prompt      bool   `json:"prompt"`      //提示音
	DownloadDir string `json:"downloadDir"` //下载地址
}

func (this *cfg) Dir() string {
	if len(this.DownloadDir) == 0 {
		return "./"
	}
	return this.DownloadDir
}

func (this *cfg) Json() string {
	bs, _ := json.Marshal(this)
	return string(bs)
}

func (this *cfg) Save() error {
	return oss.New(cfgPath, this.Json())
}
