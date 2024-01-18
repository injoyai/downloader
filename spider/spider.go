package spider

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	ByID              = "id"
	ByXPATH           = "xpath"
	ByLinkText        = "link text"
	ByPartialLinkText = "partial link text"
	ByName            = "name"
	ByTagName         = "tag name"
	ByClassName       = "class name"
	ByCSSSelector     = "css selector"
)

type Entity struct {
	system        string //当前运行系统,linux,windows
	showWindow    bool   //显示窗口
	showImg       bool   //显示图片
	seleniumPath  string //chromedriver路径
	seleniumPort  int    //selenium端口
	seleniumDebug bool   //selenium调试模式
	retry         uint   //重试次数
	try           uint   //运行次数
}

// Retry 重试机制
func (this *Entity) Retry(n uint) *Entity {
	this.retry = n
	return this
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
	this.seleniumDebug = !(len(b) > 0 && !b[0])
	return this
}

// Run 执行,记得保留加载时间
func (this *Entity) Run(fn func(i IPage)) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			log.Println("[错误]", err)
			if this.retry > this.try {
				err = this.Run(fn)
			}
		}
		//log.Println("[信息] 爬虫执行结束...")
		//<-time.After(time.Second * 20)
	}()

	//如果seleniumServer没有启动，就启动一个seleniumServer所需要的参数，可以为空，示例请参见https://github.com/tebeka/selenium/blob/master/example_test.go
	opts := []selenium.ServiceOption{}
	selenium.SetDebug(this.seleniumDebug)
	service, err := selenium.NewChromeDriverService(this.seleniumPath, this.seleniumPort, opts...)
	if nil != err {
		return err
	}
	//注意这里，server关闭之后，chrome窗口也会关闭
	defer service.Stop()

	//链接本地的浏览器 chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	//禁止图片加载，加快渲染速度
	pref := map[string]interface{}{}
	if !this.showImg || !this.showWindow {
		pref["profile.managed_default_content_settings.images"] = 2
	}
	// 模拟user-agent，防反爬
	arg := []string{"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}
	if this.system == "linux" || !this.showWindow {
		arg = append(arg, "--headless")
	}
	//设置浏览器参数
	caps.AddChrome(chrome.Capabilities{
		Path:  "./browser/chrome/chrome.exe",
		Prefs: pref,
		Args:  arg,
	})
	// 调起chrome浏览器
	web, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", this.seleniumPort))
	if err != nil {
		return err
	}

	//===========================================

	this.try++
	fn(Page{WebDriver: web})
	return
}

// New
// 新建实例需要下载chromedriver
// 查看浏览器版本Chrome://version
// http://chromedriver.storage.googleapis.com/index.html
func New(path string) *Entity {
	return &Entity{
		system:        runtime.GOOS,
		showWindow:    true,
		showImg:       true,
		seleniumPath:  path,
		seleniumPort:  20165,
		seleniumDebug: false,
	}
}

type IPage interface {
	Open(string) Page
	Exit(interface{})
	Wait(duration time.Duration) Page
	WaitMin(int) Page
	WaitSec(...int) Page
}

type Page struct {
	selenium.WebDriver
	other
}

func (this Page) Wait(t time.Duration) Page {
	this.other.Wait(t)
	return this
}

func (this Page) WaitSec(n ...int) Page {
	this.other.WaitSec(n...)
	return this
}

func (this Page) WaitMin(n int) Page {
	this.other.WaitMin(n)
	return this
}

// 返回页面数据
func (this Page) String() string {
	return this.Text()
}

// Text 返回页面数据
func (this Page) Text() string {
	s, err := this.PageSource()
	this.setErr(err)
	return s
}

// New 刷新页面,刷新页面会清空表单!!
func (this Page) New() Page {
	return this.Refresh()
}

// Refresh 刷新页面,刷新页面会清空表单!!
func (this Page) Refresh() Page {
	this.setErr(this.WebDriver.Refresh())
	return this
}

// Open 打开网页
func (this Page) Open(url string) Page {
	this.setErr(this.WebDriver.Get(url))
	return this
}

// FindXPaths 查找所有XPath
func (this Page) FindXPaths(path string) []Element {
	return this.Finds(ByXPATH, path)
}

// FindSelect 查找所有XPath
func (this Page) FindSelect(path string) Element {
	return this.Find(ByCSSSelector, path)
}

// FindSelects 查找所有XPath
func (this Page) FindSelects(path string) []Element {
	return this.Finds(ByCSSSelector, path)
}

// FindXPath 查找所有XPath
func (this Page) FindXPath(path string) Element {
	return this.Find(ByXPATH, path)
}

// Find 查找所有XPath
func (this Page) Find(by, val string) Element {
	v, err := this.FindElement(by, val)
	this.setErr(err)
	return Element{
		WebElement: v,
		Page:       this,
	}
}

// Finds 查找所有XPath
func (this Page) Finds(by, val string) []Element {
	list, err := this.FindElements(by, val)
	this.setErr(err)
	x := []Element{}
	for _, v := range list {
		x = append(x, Element{
			WebElement: v,
			Page:       this,
		})
	}
	return x
}

type Element struct {
	selenium.WebElement
	Page
}

func (this Element) Wait(t time.Duration) Element {
	this.other.Wait(t)
	return this
}

func (this Element) WaitSec(n ...int) Element {
	this.other.WaitSec(n...)
	return this
}

func (this Element) WaitMin(n int) Element {
	this.other.WaitMin(n)
	return this
}

func (this Element) String() string {
	return this.Text()
}

func (this Element) Text() string {
	s, err := this.WebElement.Text()
	this.setErr(err)
	return s
}

func (this Element) Click() Element {
	this.setErr(this.WebElement.Click())
	return this
}

func (this Element) Write(s string) Element {
	this.setErr(this.WebElement.SendKeys(s))
	return this
}

// Submit 提交
func (this Element) Submit() Element {
	this.setErr(this.WebElement.Submit())
	return this
}

type other struct {
	Error error
}

// Wait 等待
func (this other) Wait(t time.Duration) other {
	time.Sleep(t)
	return this
}

// WaitSec 等待,秒
func (this other) WaitSec(n ...int) {
	s := 1
	if len(n) > 0 {
		s = n[0]
	}
	time.Sleep(time.Second * time.Duration(s))
}

// WaitMin 等待,分钟
func (this other) WaitMin(n int) {
	time.Sleep(time.Minute * time.Duration(n))
}

func (this other) setErr(err error) other {
	if this.Error == nil && err != nil {
		this.Error = err
		this.Exit(err)
	}
	return this
}

func (this other) Exit(i interface{}) {
	panic(i)
}
