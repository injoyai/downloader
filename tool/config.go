package tool

import (
	"encoding/json"
	"io/ioutil"
)

const (
	cfgPath = "./config/config.json"
)

var Cfg = func() *cfg {
	bs, _ := ioutil.ReadFile(cfgPath)
	data := new(cfg)
	json.Unmarshal(bs, data)
	return data
}()

type cfg struct {
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
	return NewFile(cfgPath, this.Json())
}
