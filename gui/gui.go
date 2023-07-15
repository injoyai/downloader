package gui

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/download"
	"github.com/injoyai/downloader/download/m3u8"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/downloader/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/tebeka/selenium"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func New() error {
	return lorca.Run(&lorca.Config{
		Width:  600,
		Height: 390,
		Html:   "./gui/index.html",
	}, func(app lorca.APP) error {

		app.SetValueByID("done_voice", conv.String(tool.Cfg.Prompt))
		app.SetValueByID("download_dir", conv.String(tool.Cfg.DownloadDir))

		enable := chans.NewRerun(func(ctx context.Context) {
			downloadAddr := strings.TrimSpace(conv.String(app.GetValueByID("download_addr")))
			downloadDir := conv.String(app.GetValueByID("download_dir"))
			downloadName := conv.String(app.GetValueByID("download_name"))
			proxyAddr := conv.String(app.GetValueByID("proxy_addr"))
			doneVoice := conv.Bool(app.GetValueByID("done_voice"))

			logs.PrintErr(func() (err error) {
				if len(proxyAddr) > 0 {
					http.DefaultClient.SetProxy(proxyAddr)
				}
				defer g.Recover(&err, true)

				if downloadDir != tool.Cfg.Dir() || doneVoice != tool.Cfg.Prompt {
					tool.Cfg.DownloadDir = downloadDir
					tool.Cfg.Prompt = doneVoice
					tool.Cfg.Save()
				}

				// 不存在则生成保存的文件夹
				g.PanicErr(os.MkdirAll(tool.Cfg.Dir(), 0777))

				if len(downloadAddr) == 0 {
					return errors.New("无效下载地址")
				}

				app.SetValueByID("log", downloadAddr)
				urls, err := findUrl(downloadAddr)
				g.PanicErr(err)
				if len(urls) == 0 {
					app.SetValueByID("log", "没有找到资源")
					return
				}

				defer func() {
					if tool.Cfg.Prompt {
						v, _ := voice.NewLocal(nil)
						v.Call(&voice.Message{
							TemplateID: "",
							Phone:      "",
							Param:      "叮咚. 你的视频已下载完成",
						})
					}
				}()

				list := make([]string, len(urls))
				for i, url := range urls {
					func(i int, url, filename string) (err error) {
						defer func() {
							if err != nil {
								app.SetValueByID("log", err.Error())
							} else {
								app.SetValueByID("log", strings.Join(list, "\n"))
							}
						}()
						start := time.Now()
						list[i] = url
						app.SetValueByID("log", strings.Join(list, "\n"))
						l, err := m3u8.NewTask(url)
						if err != nil {
							list[i] = err.Error()
							app.SetValueByID("log", err.Error())
							return err
						}
						if len(filename) == 0 {
							filename = l.Filename()
						} else if !strings.Contains(filename, ".") {
							filename += "_" + strconv.Itoa(i) + filepath.Ext(l.Filename())
						}

						f, err := os.Create(downloadDir + filename)
						if err != nil {
							list[i] = err.Error()
							app.SetValueByID("log", err.Error())
							return err
						}

						total := float64(l.Len())
						current := uint32(0)

						errs := download.New(&download.Option{
							Limit: 20,
						}).Run(l, f, func() {
							value := atomic.AddUint32(&current, 1)
							rate := (float64(value) / total) * 100
							app.SetValueByID("bar", int(rate))
							app.SetValueByID("log", fmt.Sprintf("%0.1f%%", rate))
						})
						f.Close()
						if len(errs) > 0 {
							list[i] = errs[0].Error()
						} else {
							list[i] = "下载完成,用时" + time.Now().Sub(start).String()
						}
						return nil
					}(i, url, downloadName)
				}
				return nil
			}())
		})

		return app.Bind("run", func() {
			running := app.GetValueByID("download") == "开始下载"
			app.SetValueByID("download", conv.SelectString(running, "停止下载", "开始下载"))
			defer app.SetValueByID("download", "开始下载")
			enable.Enable(running)
		})

	})
}

func findUrl(u string) ([]string, error) {

	urls := []string(nil)
	if strings.Contains(u, ".m3u8") {
		return []string{u}, nil
	}

	if !strings.Contains(u, "http") {
		return nil, errors.New("invalid url")
	}
	if err := spider.New("./chromedriver.exe").ShowWindow(false).ShowImg(false).Run(func(i spider.IPage) {
		p := i.Open(u)
		p.WaitSec(3)

		switch {
		case strings.Contains(u, "91pron"): //处理91pron
			list := regexp.MustCompile(`VID=[0-9]+`).FindAllString(p.String(), -1)
			for _, v := range list {
				num := v[4:]
				urls = append(urls, fmt.Sprintf("https://cdn77.91p49.com/m3u8/%s/%s.m3u8", num, num))
			}
			if len(list) > 0 {
				return
			}
		}

		urls = m3u8.RegexpAll(p.String())
		iframes, err := p.FindElements(selenium.ByCSSSelector, "iframe")
		g.PanicErr(err)
		for _, v := range iframes {
			g.PanicErr(p.SwitchFrame(v))
			urls = append(urls, m3u8.RegexpAll(p.String())...)
			g.PanicErr(p.SwitchFrame(nil))
		}

	}); err != nil {
		return nil, err
	}

	{ //去除重复地址
		m := make(map[string]string)
		for _, v := range urls {
			u, err := url.Parse(v)
			if err == nil {
				m[u.Path] = v
			}
		}
		urls = []string{}
		for _, m3u8Url := range m {
			urls = append(urls, m3u8Url)
		}
	}

	{ //特殊处理网站
		for i, v := range urls {
			if strings.Contains(v, `//test.`) {
				host := str.CropLast(v, "/")
				bs, _ := http.GetBytes(host)
				for _, s := range regexp.MustCompile(`>(.*?)\.m3u8<`).FindAllString(string(bs), -1) {
					s = str.CropFirst(s, ">", false)
					s = str.CropLast(s, "<", false)
					if filepath.Base(v) != s {
						urls[i] = host + s
						break
					}
				}

			}
		}
	}

	return urls, nil
}
