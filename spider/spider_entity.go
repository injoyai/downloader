package spider

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/injoyai/selenium"
	"github.com/injoyai/selenium/chrome"
)

const (
	ByID              = selenium.ByID
	ByXPATH           = selenium.ByXPATH
	ByLinkText        = selenium.ByLinkText
	ByPartialLinkText = selenium.ByPartialLinkText
	ByName            = selenium.ByName
	ByTagName         = selenium.ByTagName
	ByClassName       = selenium.ByClassName
	ByCSSSelector     = selenium.ByCSSSelector
)

type Browser string

const (
	Chrome  Browser = "chrome"
	Firefox Browser = "firefox"
	Opera   Browser = "opera"
)

// New
// 新建实例需要下载chromedriver
// 查看浏览器版本Chrome://version
// http://chromedriver.storage.googleapis.com/index.html
// https://www.chromedownloads.net/chrome64win/
func New(browserPath, driverPath string, option ...Option) *Entity {
	e := &Entity{
		showWindow:    oss.IsWindows(),
		showImg:       true,
		browser:       Chrome,
		browserPath:   browserPath,
		driverPath:    driverPath,
		seleniumPort:  20165,
		seleniumDebug: false,
		userAgent:     http.UserAgentDefault,
		retry:         1,
	}
	for _, v := range option {
		v(e)
	}
	return e
}

type Option func(e *Entity)

type Entity struct {
	showWindow    bool    //显示窗口
	showImg       bool    //显示图片
	browser       Browser //浏览器
	browserPath   string  //浏览器目录
	driverPath    string  //chromedriver路径
	seleniumPort  int     //selenium端口
	seleniumDebug bool    //selenium调试模式
	userAgent     string  //User-Agent
	retry         uint    //重试次数
}

// SetRetry 设置重试次数
func (this *Entity) SetRetry(n uint) *Entity {
	this.retry = n
	return this
}

// SetBrowser 设置浏览器,目前只支持chrome
func (this *Entity) SetBrowser(b Browser) *Entity {
	this.browser = b
	return this
}

// SetBrowserPath 设置浏览器目录
func (this *Entity) SetBrowserPath(p string) *Entity {
	this.browserPath = p
	return this
}

// SetUserAgent 设置UserAgent
func (this *Entity) SetUserAgent(ua string) *Entity {
	this.userAgent = ua
	return this
}

// SetUserAgentDefault 设置UserAgent到默认值
func (this *Entity) SetUserAgentDefault() *Entity {
	return this.SetUserAgent(http.UserAgentDefault)
}

// SetUserAgentRand 设置随机UserAgent
func (this *Entity) SetUserAgentRand() *Entity {
	idx := g.RandInt(0, len(http.UserAgentList)-1)
	return this.SetUserAgent(http.UserAgentList[idx])
}

// ShowWindow 显示窗口linux系统无效
func (this *Entity) ShowWindow(b ...bool) *Entity {
	this.showWindow = !(len(b) > 0 && !b[0])
	return this
}

// ShowImg 是否加载图片
func (this *Entity) ShowImg(b ...bool) *Entity {
	this.showImg = !(len(b) > 0 && !b[0])
	return this
}

// SetPort 设置端口
func (this *Entity) SetPort(port int) *Entity {
	this.seleniumPort = port
	return this
}

// Debug 是否打印日志
func (this *Entity) Debug(b ...bool) *Entity {
	this.seleniumDebug = len(b) == 0 || b[0]
	return this
}

// Run 执行,记得保留加载时间
func (this *Entity) Run(f func(w *WebDriver) error, option ...selenium.ServiceOption) error {

	selenium.SetDebug(this.seleniumDebug)
	serviceOption := []selenium.ServiceOption{
		selenium.Output(logs.DefaultErr),
	}
	serviceOption = append(serviceOption, option...)
	//新建seleniumServer
	service, err := selenium.NewChromeDriverService(
		this.driverPath,
		this.seleniumPort,
		serviceOption...,
	)
	if nil != err {
		return err
	}
	defer service.Stop()

	//链接本地的浏览器 chrome
	caps := selenium.Capabilities{"browserName": string(Chrome)}
	//设置浏览器参数
	caps.AddChrome(chrome.Capabilities{
		Path: this.browserPath,
		Prefs: map[string]interface{}{
			//是否禁止图片加载，加快渲染速度
			"profile.managed_default_content_settings.images": conv.SelectInt(this.showWindow && this.showImg, 1, 2),
		},
		Args: func() []string {
			list := []string{
				"--user-agent=" + this.userAgent,
				//"--single-process",
				//"--no-sandbox", //关闭沙盒模式
				//"--incognito",  //无痕模式
			}
			if !oss.IsWindows() || !this.showWindow {
				list = append(list, "--headless")
			}
			return list
		}(),
	})

	// 调起浏览器
	web, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", this.seleniumPort))
	if err != nil {
		return err
	}
	defer web.Close()

	return g.Retry(func() error { return f(&WebDriver{web}) }, this.retry)
}
