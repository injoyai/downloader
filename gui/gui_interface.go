package gui

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/download"
	"github.com/injoyai/downloader/download/m3u8"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/lorca"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Interface interface {
	Set(key string, value string) error //设置属性
	Get(key string) (string, error)     //获取属性
	SetLog(value string)                //设置日志
	SetDownload(enable bool)            //设置下载开始/结束
}

type gui struct {
	lorca.APP
}

func (this *gui) Set(key string, value string) {
	this.APP.SetValueByID(key, value)
}

func (this *gui) Get(key string) string {
	return this.APP.GetValueByID(key)
}

func (this *gui) SetLog(value string) {
	this.Set("log", value)
}

func (this *gui) SetDownload(enable bool) {
	if enable {
		this.Set("download", "停止下载")
	} else {
		this.Set("download", "下载")
	}
}

func New2() error {
	return lorca.Run(&lorca.Config{
		Width:  600,
		Height: 442,
		Html:   html,
	}, func(app lorca.APP) error {

		app.SetValueByID("download_addr", cfg.GetString("download_addr"))
		app.SetValueByID("download_dir", cfg.GetString("download_dir", "./"))
		app.Eval(fmt.Sprintf("document.getElementById('proxy_addr').checked=%v", cfg.GetBool("proxy_addr")))
		app.SetValueByID("proxy_addr", cfg.GetString("proxy_addr", "localhost:1081"))
		app.Eval(fmt.Sprintf("document.getElementById('done_voice').checked=%v", cfg.GetBool("done_voice", true)))

		enable := chans.NewRerun(func(ctx context.Context) {

			app.SetValueByID("download", "停止下载")
			app.SetValueByID("log", "")

			downloadAddr := strings.TrimSpace(conv.String(app.GetValueByID("download_addr")))
			downloadDir := conv.String(app.GetValueByID("download_dir"))
			downloadName := conv.String(app.GetValueByID("download_name"))
			proxyUse := app.Eval("document.getElementById('proxy_use').checked").Bool()
			proxyAddr := conv.String(app.GetValueByID("proxy_addr"))
			doneVoice := app.Eval("document.getElementById('done_voice').checked").Bool()

			{ //处理参数,并保存到文件
				if len(downloadDir) == 0 {
					downloadDir = "./"
				}
				// 不存在则生成保存的文件夹
				oss.New(downloadDir, 0777)
				cfg.Set("download_addr", downloadAddr)
				cfg.Set("download_dir", downloadDir)
				cfg.Set("proxy_use", proxyUse)
				cfg.Set("proxy_addr", proxyAddr)
				cfg.Set("done_voice", doneVoice)
				cfg.Cover()
				//设置http代理
				http.DefaultClient.SetProxy(conv.SelectString(proxyUse, proxyAddr, ""))
			}

			func() (err error) {

				defer func() {
					if e := recover(); e != nil {
						err = errors.New(fmt.Sprint(e))
					}
					app.SetValueByID("download", "开始下载")
					if err != nil {
						app.SetValueByID("log", err.Error())
					} else if doneVoice {
						voice.Speak("叮咚. 你的视频已下载完成")
					}
				}()

				if len(downloadAddr) == 0 {
					return errors.New("无效下载地址")
				}

				app.SetValueByID("log", downloadAddr)
				urls, err := findUrl(downloadAddr)
				if err != nil {
					return err
				}
				if len(urls) == 0 {
					return errors.New("没有找到资源")
				}

				for i, url := range urls {
					func(i int, url, filename string) (err error) {
						start := time.Now()
						result := url
						app.SetValueByID("log", result)
						defer func() {
							if err != nil {
								app.SetValueByID("log", err.Error())
							} else {
								app.SetValueByID("log", result)
							}
						}()

						l, err := m3u8.NewTask(url)
						if err != nil {
							return err
						}
						if len(filename) == 0 {
							filename = l.Filename()
						} else if !strings.Contains(filename, ".") {
							filename += "_" + strconv.Itoa(i) + filepath.Ext(l.Filename())
						}

						f, err := os.Create(downloadDir + filename)
						if err != nil {
							return err
						}
						defer f.Close()

						total := float64(l.Len())
						current := uint32(0)

						errs := download.NewWithContext(ctx, &download.Option{Limit: 20}).Run(l, f, func() {
							value := atomic.AddUint32(&current, 1)
							rate := (float64(value) / total) * 100
							app.SetValueByID("bar", int(rate))
							app.SetValueByID("log", fmt.Sprintf("%0.1f%%", rate))
						})

						if len(errs) > 0 {
							return errs[0]
						}
						result = "下载完成,用时" + time.Now().Sub(start).String()
						return nil
					}(i, url, downloadName)
				}
				return nil
			}()
		})

		return app.Bind("run", func() {
			running := app.GetValueByID("download") == "开始下载"
			app.SetValueByID("download", conv.SelectString(running, "停止下载", "开始下载"))
			enable.Enable(running)
		})

	})
}
