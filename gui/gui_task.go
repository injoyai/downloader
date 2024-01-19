package gui

import (
	"context"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/goutil/task"
	"net/url"
	"path/filepath"
	"strings"
)

// getTask 根据资源地址获取任务
func getTask(u string) (task *task.Download, filename string, err error) {
	base, err := url.Parse(u)
	if err != nil {
		return nil, "", err
	}

	filename = filepath.Base(base.Path)
	filename = str.CropLast(filename, ".")
	suffix := ""

	switch true {
	case strings.Contains(u, ".mp4"):
		suffix = "mp4"
		task = NewMp4(u)
	default:
		suffix = "ts"
		task, err = m3u8.NewTask(u)
	}
	return task, filename + suffix, err
}

func NewMp4(url string) *task.Download {
	task := task.NewDownload()
	task.Append(GetBytes(func(ctx context.Context) ([]byte, error) {
		return http.GetBytes(url)
	}))
	return task
}

type GetBytes func(ctx context.Context) ([]byte, error)

func (this GetBytes) GetBytes(ctx context.Context) ([]byte, error) {
	return this(ctx)
}
