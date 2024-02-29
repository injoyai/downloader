package logic

import (
	"context"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type (
	HandlerItem func(ctx context.Context, resp *task.DownloadItemResp)
	HandlerDone func(ctx context.Context, name, source string, fn HandlerItem) error
)

func Download(ctx context.Context, op *Config) error {

	u, err := url.Parse(op.Resource)
	if err != nil {
		return err
	}

	http.DefaultClient.SetTimeout(0)
	if err := http.SetProxy(conv.SelectString(op.ProxyEnable, op.ProxyAddress, "")); err != nil {
		return err
	}

	ext := path.Ext(u.Path)
	switch ext {
	case ".m3u8":
		op.suffix = ".ts"
		err = downloadM3u8(ctx, op)

	default:
		op.suffix = ext
		err = download(ctx, op)

	}

	if err != nil {
		return err
	}

	//提示消息
	if op.NoticeEnable {
		notice.NewWindows().Publish(&notice.Message{
			Title:   "下载完成",
			Content: op.NoticeText,
		})
	}

	//播放声音
	if op.VoiceEnable {
		notice.NewVoice(nil).Speak(op.VoiceText)
	}

	return nil
}

func download(ctx context.Context, op *Config) error {

	resp := http.Get(op.Resource)
	if resp.Err() != nil {
		return resp.Err()
	}
	defer resp.Body.Close()

	b := bar.NewWithContext(ctx, resp.ContentLength)
	go b.Run()

	f, err := os.Create(op.Filename())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.CopyWithPlan(f, resp.Body, func(p *io.Plan) {
		b.Add(int64(len(p.Bytes)))
	})
	return err
}

func downloadM3u8(ctx context.Context, op *Config) error {

	resp, err := m3u8.NewResponse(op.Resource)
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

		sum := int64(0)
		current := int64(0)
		b := bar.NewWithContext(ctx, int64(len(list)))
		b.SetFormatter(func(e *bar.Format) string {
			return fmt.Sprintf("\r%s  %s  %s  %s",
				e.Bar,
				e.RateSize,
				oss.SizeString(sum),
				b.SpeedUnit("speed", current, time.Millisecond*500),
			)
		})
		go b.Run()

		//分片目录
		cacheDir := op.TempDir()

		//查看已经下载的分片
		doneName := map[string]bool{}
		oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
			if !info.IsDir() && strings.HasSuffix(info.Name(), op.suffix) {
				doneName[info.Name()] = true
			}
			return true, nil
		})

		//新建下载任务
		t := task.NewDownload()
		t.SetCoroutine(op.Coroutine)
		t.SetRetry(op.Retry)
		t.SetDoneItem(func(ctx context.Context, resp *task.DownloadItemResp) {
			if resp.Err == nil {
				//保存分片到文件夹,5位长度,最大99999分片,大于99999视频会乱序
				filename := fmt.Sprintf("%05d"+op.suffix, resp.Index)
				filename = filepath.Join(cacheDir, filename)
				g.Retry(func() error { return oss.New(filename, resp.Bytes) }, 3)
			}
			current = resp.GetSize()
			sum += current
			b.Add(1)
		})
		for i, v := range list {
			filename := fmt.Sprintf("%05d"+op.suffix, i)
			if doneName[filename] {
				//过滤已经下载过的分片
				b.Add(1)
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
		totalFile, err := os.Create(op.Filename())
		if err != nil {
			return err
		}

		//合并视频
		g.Retry(func() error {
			return oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
				if !info.IsDir() && strings.HasSuffix(info.Name(), op.suffix) {
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
		totalFile.Close()

		//删除文件夹和分片视频
		oss.DelDir(cacheDir)

		break

	}

	return nil
}

type Config struct {
	Resource     string
	Dir          string
	Name         string
	suffix       string
	Retry        uint
	Coroutine    uint
	ProxyEnable  bool
	ProxyAddress string
	NoticeEnable bool
	NoticeText   string
	VoiceEnable  bool
	VoiceText    string
}

func (this *Config) GetName() string {
	if len(this.Name) == 0 {
		u, err := url.Parse(this.Resource)
		if err == nil {
			this.Name = strings.Split(path.Base(u.Path), ".")[0]
		}
	}
	if len(this.Name) == 0 {
		this.Name = time.Now().Format("20060102150405")
	}
	return this.Name
}

func (this *Config) Filename() string {
	return filepath.Join(this.Dir, this.GetName()+this.suffix)
}

func (this *Config) TempDir() string {
	return filepath.Join(this.Dir, this.GetName())
}
