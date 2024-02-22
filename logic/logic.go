package logic

import (
	"context"
	"fmt"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/task"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	HandlerInfo func(i *Info) *Info
	HandlerItem func(ctx context.Context, resp *task.DownloadItemResp)
	HandlerDone func(ctx context.Context, name, source string, fn HandlerItem) error
)

func Download(ctx context.Context, dir, source string, f1 HandlerInfo, fn HandlerItem) error {
	return downloadM3u8(ctx, dir, source, f1, fn)
}

func downloadM3u8(ctx context.Context, dir, source string, f1 HandlerInfo, fn HandlerItem) error {

	resp, err := m3u8.NewResponse(source)
	if err != nil {
		return err
	}

	lists, err := resp.List()
	if err != nil {
		return err
	}

	if len(lists) == 0 {
		return nil
	}

	cacheDir := filepath.Join(dir, resp.Name())

	for _, list := range lists {

		//查看已经下载的分片
		doneName := map[string]bool{}
		oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) bool {
			doneName[info.Name()] = true
			return true
		})

		config := &Info{
			Total:   int64(len(lists[0])),
			Current: int64(len(doneName)),
			Name:    resp.Name(),
			Config:  DefaultConfig,
		}
		config = f1(config)

		//新建下载任务
		t := task.NewDownload()
		t.SetCoroutine(config.Coroutine)
		t.SetRetry(config.Retry)
		t.SetDoneItem(func(ctx context.Context, resp *task.DownloadItemResp) {
			if resp.Err != nil {
				//保存分片到文件夹
				g.Retry(func() error { return oss.New(fmt.Sprintf("%04d"+config.Suffix, resp.Index), resp.Bytes) }, 3)
			}
			fn(ctx, resp)
		})
		for i, v := range list {
			if doneName[strconv.Itoa(i)] {
				//过滤已经下载过的分片
				continue
			}
			//继续下载没有下载过的分片
			t.Append(v)
		}

		doneResp := t.Download(ctx)
		if doneResp.Err != nil {
			return doneResp.Err
		}
		//合并视频,删除分片等信息
		totalFile, err := os.Create(cacheDir + config.Suffix)
		if err != nil {
			return err
		}
		g.Retry(func() error {
			return oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) bool {
				if !info.IsDir() && strings.HasSuffix(info.Name(), config.Suffix) {
					f, err := os.Open(filepath.Join(cacheDir, info.Name()))
					if err != nil {
						return false
					}
					defer f.Close()
					_, err = io.Copy(totalFile, f)
					return err == nil
				}
				return true
			})
		}, 3)
		//删除文件夹
		oss.DelDir(cacheDir)

	}

	return nil
}

var (
	DefaultConfig = &Config{
		Retry:     3,
		Coroutine: 20,
		Dir:       "./",
		Suffix:    ".ts",
	}
)

type Config struct {
	Retry     uint
	Coroutine uint
	Dir       string
	Suffix    string
}

type Info struct {
	Total   int64
	Current int64
	Name    string
	*Config
}
