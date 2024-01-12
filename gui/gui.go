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
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/goutil/str/bar"
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
	CoroutineNum int    //协程数量
}

func (this *Config) deepFind(p spider.Page) ([]string, error) {
	urls := m3u8.RegexpAll(p.String())
	iframes, err := p.FindElements(selenium.ByCSSSelector, "iframe")
	if err != nil {
		return nil, err
	}
	for _, v := range iframes {
		if err := p.SwitchFrame(v); err != nil {
			logs.Err(err)
			return nil, err
		}
		ls, err := this.deepFind(p)
		if err != nil {
			return nil, err
		}
		urls = append(urls, ls...)
		if err := p.SwitchFrame(nil); err != nil {
			logs.Err(err)
			return nil, err
		}
	}
	return urls, nil
}

// FindUrl 通过资源地址获取到下载连接
func (this *Config) FindUrl() (urls []string, err error) {

	u := this.DownloadAddr

	if strings.Contains(u, ".m3u8") {
		return []string{u}, nil
	}

	if !strings.Contains(u, "http") {
		return nil, errors.New("无效资源地址")
	}

	logs.Debug("开始爬取...")
	if err := spider.New("./chromedriver.exe").ShowWindow(false).ShowImg(false).Run(func(i spider.IPage) {
		p := i.Open(u)
		p.WaitSec(3)

		//正则匹配数据,包括iframe
		urls, err = this.deepFind(p)
		g.PanicErr(err)

		//去除转义符
		for idx, v := range urls {
			urls[idx] = strings.ReplaceAll(v, `\/`, "/")
		}

		switch {

		case strings.Contains(u, "91pron"):

			//特殊处理91pron
			list := regexp.MustCompile(`VID=[0-9]+`).FindAllString(p.String(), -1)
			for _, v := range list {
				num := v[4:]
				urls = append(urls, fmt.Sprintf("https://cdn77.91p49.com/m3u8/%s/%s.m3u8", num, num))
			}

		case strings.Contains(u, "bedroom.uhnmon.com") || strings.Contains(u, "/51cg") || strings.Contains(u, "hy9hz1.xxousm.com"):

			//特殊处理51cg
			for idx, v := range urls {
				urls[idx] = v + "&v=3&time=0"
			}

		default:

			//特殊处理网站,忘记是啥网站了
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

	}); err != nil {
		return nil, err
	}
	logs.Debug("爬取成功...")

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

	if len(urls) == 0 {
		return nil, errors.New("没有找到资源")
	}
	logs.Debug("资源地址: ", urls)

	return urls, nil
}

// createFile 新建文件
func (this *Config) createFile(idx int, urlFilename string) (*os.File, error) {
	filename := conv.SelectString(len(this.DownloadName) == 0, urlFilename, this.DownloadName+"_"+strconv.Itoa(idx)+filepath.Ext(urlFilename))
	logs.Debug("文件名称: ", filename)
	return os.Create(this.DownloadDir + filename)
}

// Download 下载
func (this *Config) Download(ctx context.Context, gui *gui, idx int, url string) (size int64, err error) {

	logs.Debug("----------------------------------------------------------------------------------------------------")
	logs.Debug("资源地址: ", url)

	task, filename, err := m3u8.NewTask(url)
	if err != nil {
		return 0, err
	}

	logs.Debug("分片数量: ", task.Len())
	logs.Debug("协程数量: ", this.CoroutineNum)

	f, err := this.createFile(idx, filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	done := make(chan struct{}, 1)
	task.SetWriter(f)
	task.SetLimit(20)
	task.SetRetry(3)
	task.SetDoneAll(func() {
		done <- struct{}{}
	})
	current := uint32(0)
	errs := []error(nil)
	task.SetDoneItem(func(i int, err error) {
		value := atomic.AddUint32(&current, 1)
		rate := (float64(value) / float64(task.Len())) * 100
		gui.SetBar(rate)
		gui.SetLog(fmt.Sprintf("%0.1f%%", rate))
		if err != nil {
			errs = append(errs, err)
		}
	})

	download.NewWithContext(ctx).Append(task)
	<-done

	fs, err := f.Stat()
	if err != nil {
		return 0, append(errs, nil)[0]
	}

	return fs.Size(), append(errs, nil)[0]
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
				size, err := config.Download(ctx, gui, i, url)
				gui.SetLog(conv.SelectString(err == nil, "下载成功,用时"+time.Now().Sub(start).String(), "下载失败: "+conv.String(err)))
				logs.Debug("下载结果: ", conv.New(err).String("成功"))
				logs.Debug("下载用时: ", time.Now().Sub(start).String())
				fSize, unit := bar.ToB(size)
				logs.Debugf("文件大小: %0.2f%s\n\n", fSize, unit)
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
