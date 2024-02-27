package gui

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/selenium"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed index.html
var html string

type gui struct {
	lorca.APP
	*cache.File
	driverPath  string
	browserPath string
}

func (this *gui) Set(key string, value interface{}) {
	this.APP.SetValueByID(key, value)
}

func (this *gui) Get(key string) string {
	return this.APP.GetValueByID(key)
}

func (this *gui) SetLog(value interface{}) {
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

// openSettings 打开设置
func (this *gui) openSettings() {
	this.Eval("document.getElementById('settingsModal').style.display = 'block';")
}

// applySettings 确认设备
func (this *gui) applySettings() {
	c := &Config{
		DownloadDir:  conv.String(this.APP.GetValueByID("download_dir")),
		DownloadName: conv.String(this.APP.GetValueByID("download_name")),
		ProxyEnable:  this.APP.Eval("document.getElementById('proxy_use').checked").Bool(),
		ProxyAddr:    conv.String(this.APP.GetValueByID("proxy_addr")),
		NoticeEnable: this.APP.Eval("document.getElementById('notice_enable').checked").Bool(),
		NoticeText:   this.APP.GetValueByID("notice_text"),
		RetryNum:     conv.Uint(this.APP.GetValueByID("retry_num")),
		CoroutineNum: conv.Uint(this.APP.GetValueByID("coroutine_num")),
	}
	if len(c.DownloadDir) == 0 {
		c.DownloadDir = "./"
	}
	c.DownloadDir = strings.ReplaceAll(c.DownloadDir, "\\", "/")
	if c.CoroutineNum == 0 {
		c.CoroutineNum = 20
	}
	if c.RetryNum == 0 {
		c.RetryNum = 3
	}
	oss.New(c.DownloadDir, 0777)
	this.File.Set("download_dir", c.DownloadDir)
	this.File.Set("proxy_enable", c.ProxyEnable)
	this.File.Set("proxy_addr", c.ProxyAddr)
	this.File.Set("notice_enable", c.NoticeEnable)
	this.File.Set("notice_text", c.NoticeText)
	this.File.Set("retry_num", c.RetryNum)
	this.File.Set("coroutine_num", c.CoroutineNum)
	this.File.Cover()
	this.closeSettings()
}

// closeSettings 关闭设置
func (this *gui) closeSettings() {
	this.Eval("document.getElementById('settingsModal').style.display = 'none';")
}

func (this *gui) loadResources() {

}

func (this *gui) findUrl(ctx context.Context) {

	logs.PrintErr(
		spider.New(
			this.browserPath,
			this.driverPath,
			func(e *spider.Entity) {
				e.SetRetry(1)
			},
		).Run(func(w *selenium.WebDriver) error {

			logs.Debug("打开浏览器")
			w.Get("https://www.google.com")

			handler, err := w.CurrentWindowHandle()
			if err != nil {
				return err
			}
			handlerMap := maps.NewSafe()
			handlerMap.Set(handler, true)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go this.findUrl2(ctx, wg, handlerMap, w, time.Second*2)
			wg.Wait()

			return nil
		}))

}

func (this *gui) findUrl2(ctx context.Context, wg *sync.WaitGroup, handlerMap *maps.Safe, w *selenium.WebDriver, interval time.Duration) error {
	defer wg.Done()

	timer := time.NewTimer(interval)
	for {
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			logs.Debug("关闭标签页")
			return errors.New("关闭标签页")
		case <-timer.C:

			logs.Debug("寻找资源")

			logs.Debug(w.CurrentWindowHandle())
			logs.Debug(w.SessionID())

			handlerList, err := w.WindowHandles()
			if err != nil {
				return err
			}

			for _, handler := range handlerList {
				//if i != len(handlerList)-1 {
				//	continue
				//}
				handlerMap.GetOrSetByHandler(handler, func() (interface{}, error) {
					logs.Debug(handler)
					//w2 := w.NewSeesion(handler)
					//w2.SwitchWindow()
					wg.Add(1)
					go this.findUrl2(ctx, wg, handlerMap, w, interval)
					return time.Now(), nil
				})
			}

			title, err := w.Title()
			if err != nil {
				return err
			}
			logs.Debug("标题: ", title)

			//获取hls地址
			s, err := w.Text()
			if err != nil {
				return err
			}
			m3u8List := m3u8.RegexpAll(s)
			for i, v := range m3u8List {
				m3u8List[i] = strings.ReplaceAll(v, "\\/", "/")
			}

			hrefs, err := w.FindTagAttributes("a.href")
			if err != nil {
				return err
			}
			for _, v := range hrefs {
				if m3u8.Regexp().MatchString(v) {
					m3u8List = append(m3u8List, v)
				}
			}

			for _, v := range m3u8List {
				logs.Debug(v)
			}

			//显示到GUI上
			_ = m3u8List
			this.SetLog(strings.Join(m3u8List, "\n"))

		}
	}
}

// Config 配置字段
type Config struct {
	DownloadDir  string //下载目录
	DownloadName string //下载名称
	ProxyEnable  bool   //启用代理
	ProxyAddr    string //代理地址
	NoticeEnable bool   //下载完成声音
	NoticeText   string //完成提示内容
	RetryNum     uint   //重试次数
	CoroutineNum uint   //协程数量
}

// filename 新文件名称
func (this *Config) filename(urlFilename string) string {
	name := conv.SelectString(len(this.DownloadName) == 0, urlFilename, this.DownloadName)
	filename := filepath.Join(this.DownloadDir, name)
	return strings.ReplaceAll(filename, "\\", "/")
}

// Download 下载
func (this *Config) Download(ctx context.Context, gui *gui, url string) {

	//	logs.Debug("资源地址: ", url)

	tasks, filename, err := getTask(url, "")
	if err != nil {
		gui.SetLog("下载失败: " + conv.String(err))
		return
	}

	for _, t := range tasks {

		logs.Debug("----------------------------------------------------------------------------------------------------")
		logs.Debug("分片数量: ", t.Len())
		logs.Debug("协程数量: ", this.CoroutineNum)
		logs.Debug("重试次数: ", this.RetryNum)

		filename = this.filename(filename)
		logs.Debug("文件名称: ", filename)

		current := uint32(0)
		size := int64(0)
		start := time.Now()
		t.SetCoroutine(this.CoroutineNum)
		t.SetRetry(this.RetryNum)
		t.SetDoneItem(func(ctx context.Context, resp *task.DownloadItemResp) {
			value := atomic.AddUint32(&current, 1)
			size += resp.GetSize()
			rate := (float64(value) / float64(t.Len())) * 100
			gui.SetBar(rate)
			speed, speedUnit := oss.SizeUnit(size)
			speed /= time.Since(start).Seconds()
			gui.SetLog(fmt.Sprintf("%0.1f%%  %0.1f%s/s                                            %s", rate, speed, speedUnit, url))
			if resp.Err != nil {
				logs.Errf("分片(%d)下载失败: %s", resp.Index, resp.Err.Error())
			}
		})
		//resp := t.Download(ctx)
		//_, err = resp.WriteToFile(filename)

		spend := time.Now().Sub(start)
		fSize, unit := oss.SizeUnit(size)
		sizeStr := fmt.Sprintf("%0.2f%s", fSize, unit)
		spendStr := fmt.Sprintf("%0.1f%s/s", fSize/spend.Seconds(), unit)
		gui.SetLog(conv.SelectString(err == nil,
			"下载成功"+
				", 大小:"+sizeStr+
				", 用时:"+time.Now().Sub(start).String()+
				", 速度:"+spendStr+
				"      文件位置:"+filename,
			"下载失败: "+conv.String(err)))
		logs.Debug("下载结果: ", conv.New(err).String("成功"))
		logs.Debug("下载用时: ", spend.String())
		logs.Debug("文件大小: ", sizeStr)
		logs.Debug("平均速度: ", spendStr)
		logs.Debug("----------------------------------------------------------------------------------------------------")

	}

}

func New(configPath, driverPath, browserPath string) error {
	return lorca.Run(&lorca.Config{
		Width:  610,
		Height: 500,
		Html:   html,
	}, func(app lorca.APP) error {

		logs.Debug(driverPath)
		logs.Debug(browserPath)
		gui := &gui{
			APP:         app,
			File:        cache.NewFile(configPath),
			driverPath:  driverPath,
			browserPath: browserPath,
		}

		//设置配置信息到gui
		gui.SetConfig()

		gui.Bind("openSettings", gui.openSettings)
		gui.Bind("applySettings", gui.applySettings)
		gui.Bind("closeSettings", gui.closeSettings)

		enable := chans.NewRerun(gui.findUrl)

		return gui.Bind("openBrowser", func() {
			logs.Debug(gui.Get("open_browser"))
			opened := gui.Get("open_browser") == "close"
			gui.Set("open_browser", conv.SelectString(opened, "close", "open"))
			enable.Enable(opened)
		})

	})
}
