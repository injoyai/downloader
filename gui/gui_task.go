package gui

import (
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	Mp4  = "mp4"
	M3u8 = "m3u8"
	Ts   = "ts"
)

// getTask 根据资源地址获取任务
// @u,资源地址
// @resourceType,资源类型m3u8或者mp4等,为空表示未知,需要自行判断
func getTask(u string, resourceType string) (task []*task.Download, filename string, err error) {
	base, err := url.Parse(u)
	if err != nil {
		logs.Err("网址解析错误: ", err)
		return nil, "", err
	}

	suffix := ""
	switch true {
	case resourceType == M3u8 || (len(resourceType) == 0 && strings.HasSuffix(base.Path, M3u8)):
		suffix = Ts
		task, err = m3u8.NewTask(u)

	//case resourceType == Mp4 || (len(resourceType) == 0 && strings.HasSuffix(base.Path, Mp4)):
	//	suffix = Mp4
	//	task = mp4.NewTask(u)

	default:
		suffix = Ts
		task, err = m3u8.NewTask(u)

	}

	logs.Debug("base.Path: ", base.Path)
	filename = filepath.Base(strings.ReplaceAll(base.Path, "//", "/"))
	filename = str.CropLast(filename, ".")
	return task, filename + suffix, err
}
