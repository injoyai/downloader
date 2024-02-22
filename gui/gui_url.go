package gui

import (
	"errors"
	"fmt"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/selenium"
	"net/url"
	"regexp"
	"strings"
)

var (
	Regexp = regexp.MustCompile(`(http:|https:)[a-zA-Z0-9\\/=_\-.:,%&]+\.(m3u8)([?a-zA-Z0-9/=_\-.*%&]{0,})`)
)

func RegexpAll(s string) []string {
	return Regexp.FindAllString(s, -1)
}

type Element interface {
	PageSource() (string, error)
	FindElements(by, value string) ([]selenium.WebElement, error)
}

func (this *Config) deepFindElement(e selenium.WebElement) ([]string, error) {
	text, err := e.Text()
	if err != nil {
		return nil, err
	}
	urls := RegexpAll(text)
	iframes, err := e.FindElements(selenium.ByCSSSelector, "iframe")
	if err != nil {
		return nil, err
	}
	for _, v := range iframes {
		ls, err := this.deepFindElement(v)
		if err != nil {
			return nil, err
		}
		urls = append(urls, ls...)
	}
	return urls, nil
}

func (this *Config) deepFind(w Element) ([]string, error) {
	text, err := w.PageSource()
	if err != nil {
		return nil, err
	}
	urls := RegexpAll(text)
	iframes, err := w.FindElements(selenium.ByCSSSelector, "iframe")
	if err != nil {
		return nil, err
	}
	for _, v := range iframes {
		ls, err := this.deepFindElement(v)
		if err != nil {
			return nil, err
		}
		urls = append(urls, ls...)
	}
	//去除转义符
	for idx, v := range urls {
		urls[idx] = strings.ReplaceAll(v, `\/`, "/")
	}
	return urls, nil
}

// FindUrlWithSelenium 通过资源地址获取到下载连接
func (this *Config) FindUrlWithSelenium(driverPath, browserPath string) (urls []string, err error) {

	u := this.DownloadAddr

	if strings.Contains(u, ".m3u8") {
		return []string{u}, nil
	}

	if !strings.Contains(u, "http") {
		return nil, errors.New("无效资源地址")
	}

	fmt.Printf("\n\n")
	logs.Debug("驱动位置: ", driverPath)
	logs.Debug("浏览器位置: ", browserPath)
	logs.Debug("开始爬取: ", u)
	if err := spider.New(driverPath, browserPath).
		ShowWindow(false).ShowImg(false).Run(func(w *spider.WebDriver) (err error) {

		g.Recover(&err)

		g.PanicErr(w.Open(u))
		w.WaitSec(3)

		title, err := w.Text()
		g.PanicErr(err)

		logs.Debug("网站标题: ", title)

		//获取接口资源
		resource, err := newResource(u, title, w.WebDriver)
		logs.Debug("接口名称: ", resource.Name())

		//前置操作
		g.PanicErr(resource.Before(w.WebDriver))

		//正则匹配数据,包括iframe
		urls, err = this.deepFind(w.WebDriver)
		g.PanicErr(err)

		// 后置操作
		urls, err = resource.Deal(w.WebDriver, urls)
		g.PanicErr(err)

		return nil

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

	logs.Debug("爬取成功...")
	for _, u := range urls {
		logs.Debug("爬取地址: ", u)
	}

	//urls = []string{}
	//return

	if len(urls) == 0 {
		return nil, errors.New("没有找到资源")
	}

	return urls, nil
}
