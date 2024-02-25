package logic

import (
	"context"
	"fmt"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/task"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type (
	HandlerInfo func(i *Info) *Info
	HandlerItem func(ctx context.Context, resp *task.DownloadItemResp)
	HandlerDone func(ctx context.Context, name, source string, fn HandlerItem) error
)

func Download(ctx context.Context, source string, f1 HandlerInfo, fn HandlerItem) error {
	return downloadM3u8(ctx, source, f1, fn)
}

func downloadM3u8(ctx context.Context, source string, f1 HandlerInfo, fn HandlerItem) error {

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

	for _, list := range lists {

		//获取配置
		config := f1(&Info{
			Total:   int64(len(lists[0])),
			Current: 0,
			Name:    resp.Name(),
			Config:  DefaultConfig,
		})

		//设置代理
		if config.ProxyEnable {
			http.SetProxy(config.ProxyAddress)
		}

		//分片目录
		cacheDir := filepath.Join(config.Dir, config.Name)

		//查看已经下载的分片
		doneName := map[string]bool{}
		oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
			if !info.IsDir() && strings.HasSuffix(info.Name(), config.Suffix) {
				doneName[info.Name()] = true
			}
			return true, nil
		})

		//新建下载任务
		t := task.NewDownload()
		t.SetCoroutine(config.Coroutine)
		t.SetRetry(config.Retry)
		t.SetDoneItem(func(ctx context.Context, resp *task.DownloadItemResp) {
			if resp.Err == nil {
				//保存分片到文件夹,5位长度,最大99999分片,大于99999视频会乱序
				filename := fmt.Sprintf("%05d"+config.Suffix, resp.Index)
				filename = filepath.Join(cacheDir, filename)
				g.Retry(func() error { return oss.New(filename, resp.Bytes) }, 3)
			}
			fn(ctx, resp)
		})
		for i, v := range list {
			filename := fmt.Sprintf("%05d"+config.Suffix, i)
			if doneName[filename] {
				//过滤已经下载过的分片
				fn(ctx, &task.DownloadItemResp{
					Index: i,
				})
				continue
			}
			//继续下载没有下载过的分片
			t.Set(i, v)
		}

		//新建任务
		doneResp := t.Download(ctx)
		if doneResp.Err != nil {
			return doneResp.Err
		}
		//合并视频,删除分片等信息
		totalFile, err := os.Create(cacheDir + config.Suffix)
		if err != nil {
			return err
		}

		//合并视频
		g.Retry(func() error {
			return oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
				if !info.IsDir() && strings.HasSuffix(info.Name(), config.Suffix) {
					f, err := os.Open(filepath.Join(cacheDir, info.Name()))
					if err != nil {
						return false, err
					}
					defer f.Close()
					_, err = io.Copy(totalFile, f)
					return err == nil, err
				}
				return true, nil
			})
		}, 3)

		//删除文件夹和分片视频
		oss.DelDir(cacheDir)

		//提示消息
		if config.NoticeEnable {
			notice.NewWindows().Publish(&notice.Message{
				Title:   "下载完成",
				Content: config.NoticeText,
			})
		}

		//播放声音
		if config.VoiceEnable {
			notice.NewVoice(nil).Speak(config.VoiceText)
		}

		break

	}

	return nil
}

var (
	DefaultConfig = &Config{
		Retry:        3,
		Coroutine:    20,
		Dir:          "./",
		Suffix:       ".ts",
		ProxyEnable:  false,
		ProxyAddress: "http://127.0.0.1:1081",
		NoticeEnable: true,
		NoticeText:   "主人. 您的视频已下载结束",
		VoiceEnable:  true,
		VoiceText:    "主人. 您的视频已下载结束",
	}
)

type Config struct {
	Retry        uint
	Coroutine    uint
	Dir          string
	Suffix       string
	ProxyEnable  bool
	ProxyAddress string
	NoticeEnable bool
	NoticeText   string
	VoiceEnable  bool
	VoiceText    string
}

type Info struct {
	Total   int64
	Current int64
	Name    string
	*Config
}
