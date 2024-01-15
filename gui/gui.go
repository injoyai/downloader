package gui

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/other/download"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
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
	*cache.File
}

func (this *gui) Set(key string, value interface{}) {
	this.APP.SetValueByID(key, value)
}

func (this *gui) Get(key string) string {
	return this.APP.GetValueByID(key)
}

func (this *gui) SetLog(value string) {
	this.Set("log", value)
}

func (this *gui) SetBar(rate float64) {
	this.Set("bar", rate)
}

// SetConfig 设置配置到界面
func (this *gui) SetConfig() {
	this.APP.SetValueByID("download_addr", this.GetString("download_addr"))
	this.APP.SetValueByID("download_dir", this.GetString("download_dir", "./"))
	this.APP.Eval(fmt.Sprintf("document.getElementById('proxy_addr').checked=%v", this.GetBool("proxy_addr")))
	this.APP.SetValueByID("proxy_addr", this.GetString("proxy_addr", "localhost:1081"))
	this.APP.Eval(fmt.Sprintf("document.getElementById('done_voice').checked=%v", this.GetBool("done_voice", true)))

}

// GetConfig 获取配置,并保存
func (this *gui) GetConfig() (*Config, error) {
	c := &Config{
		DownloadAddr: strings.TrimSpace(conv.String(this.APP.GetValueByID("download_addr"))),
		DownloadDir:  conv.String(this.APP.GetValueByID("download_dir")),
		DownloadName: conv.String(this.APP.GetValueByID("download_name")),
		EnableProxy:  this.APP.Eval("document.getElementById('proxy_use').checked").Bool(),
		ProxyAddr:    conv.String(this.APP.GetValueByID("proxy_addr")),
		DoneVoice:    this.APP.Eval("document.getElementById('done_voice').checked").Bool(),
	}
	if len(c.DownloadDir) == 0 {
		c.DownloadDir = "./"
	}
	if c.CoroutineNum == 0 {
		c.CoroutineNum = 20
	}
	if c.RetryNum == 0 {
		c.RetryNum = 3
	}
	oss.New(c.DownloadDir, 0777)
	if err := this.saveConfig(c); err != nil {
		return nil, err
	}
	if len(c.DownloadAddr) == 0 {
		return nil, errors.New("无效下载地址")
	}
	this.SetLog(c.DownloadAddr)
	return c, nil
}

// saveConfig 保存配置
func (this *gui) saveConfig(cfg *Config) error {
	this.File.Set("download_addr", cfg.DownloadAddr)
	this.File.Set("download_dir", cfg.DownloadDir)
	this.File.Set("proxy_addr", cfg.ProxyAddr)
	this.File.Set("done_voice", cfg.DoneVoice)
	return this.File.Cover()
}

func (this *gui) DownloadDriver() error {
	//b:=bar.New().
	return spider.Install(func(bs []byte) {

	})
}

// Config 配置字段
type Config struct {
	DownloadAddr string //资源地址
	DownloadDir  string //下载目录
	DownloadName string //下载名称
	EnableProxy  bool   //启用代理
	ProxyAddr    string //代理地址
	DoneVoice    bool   //下载完成声音
	CoroutineNum uint   //协程数量
	RetryNum     uint   //重试次数
}

// filename 新文件名称
func (this *Config) filename(idx int, urlFilename string) string {
	return conv.SelectString(len(this.DownloadName) == 0, urlFilename, this.DownloadName+"_"+strconv.Itoa(idx)+filepath.Ext(urlFilename))
}

// Download 下载
func (this *Config) Download(ctx context.Context, gui *gui, idx int, url string) (size int64, err error) {

	logs.Debug("资源地址: ", url)

	task, filename, err := getTask(url)
	if err != nil {
		return 0, err
	}

	logs.Debug("分片数量: ", task.Len())
	logs.Debug("协程数量: ", this.CoroutineNum)
	logs.Debug("重试次数: ", this.RetryNum)

	filename = this.filename(idx, filename)
	logs.Debug("文件名称: ", filename)

	current := uint32(0)
	start := time.Now()
	task.SetDoneAllWithFile(filename)
	task.SetLimit(this.CoroutineNum)
	task.SetRetry(this.RetryNum)
	task.SetDoneItem(func(ctx context.Context, resp *download.DoneItemResp) {
		value := atomic.AddUint32(&current, 1)
		size += resp.GetSize()
		rate := (float64(value) / float64(task.Len())) * 100
		gui.SetBar(rate)
		speed, speedUnit := bar.ToB(size)
		speed /= time.Since(start).Seconds()
		gui.SetLog(fmt.Sprintf("%0.1f%%  %0.1f%s/s", rate, speed, speedUnit))
	})

	return size, task.Download(ctx)
}

func New() error {
	return lorca.Run(&lorca.Config{
		Width:  600,
		Height: 442,
		Html:   html,
	}, func(app lorca.APP) error {

		gui := &gui{
			APP:  app,
			File: cache.NewFile(oss.UserLocalDir(oss.DefaultName, "/download/config.json")),
		}

		//设置配置信息到gui
		gui.SetConfig()

		enable := chans.NewRerun(func(ctx context.Context) {

			gui.Set("download", "停止下载")
			gui.Set("bar", 0)
			defer gui.Set("download", "开始下载")

			//获取配置信息,并保存
			config, err := gui.GetConfig()
			if err != nil {
				gui.SetLog(fmt.Sprintf("%#v", err.Error()))
				return
			}

			//根据配置获取到下载地址
			urls, err := config.FindUrl()
			if err != nil {
				logs.Err(err)
				gui.SetLog(fmt.Sprintf("%#v", err.Error()))
				if strings.Contains(err.Error(), "unknown error - 33: session not created: This version of ChromeDriver only supports Chrome version") {
					gui.SetLog("浏览器和驱动不兼容,请手动删除老版本驱动chromedriver.exe,然后重启")
				}
				return
			}

			//开始下载,按顺序下载
			for i, url := range urls {
				gui.SetLog(url)
				start := time.Now()
				logs.Debug("----------------------------------------------------------------------------------------------------")
				size, err := config.Download(ctx, gui, i, url)

				spend := time.Now().Sub(start)
				fSize, unit := bar.ToB(size)
				sizeStr := fmt.Sprintf("%0.2f%s", fSize, unit)
				spendStr := fmt.Sprintf("%0.1f%s/s", fSize/spend.Seconds(), unit)
				gui.SetLog(conv.SelectString(err == nil,
					"下载成功"+
						", 大小:"+sizeStr+
						", 用时:"+time.Now().Sub(start).String()+
						", 速度:"+spendStr,
					"下载失败: "+conv.String(err)))
				logs.Debug("下载结果: ", conv.New(err).String("成功"))
				logs.Debug("下载用时: ", spend.String())
				logs.Debug("文件大小: ", sizeStr)
				logs.Debug("平均速度: ", spendStr)
				logs.Debug("----------------------------------------------------------------------------------------------------")
			}

			//播放下载完成提示音
			if config.DoneVoice {
				go voice.Speak("叮咚. 你的视频已下载结束")
			}

		})

		return gui.Bind("run", func() {
			running := gui.Get("download") == "开始下载"
			gui.Set("download", conv.SelectString(running, "停止下载", "开始下载"))
			if running {
				gui.Set("bar", 0)
				gui.SetLog("开始下载")
			}
			enable.Enable(running)
		})

	})
}
