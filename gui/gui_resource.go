package gui

import (
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/logs"
	"github.com/injoyai/selenium"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func newResource(url string, title string, w *selenium.WebDriver) (Resource, error) {
	resource := Resource(&base{})
	for _, v := range Resources {
		is, err := v.Is(url, title, w)
		if err != nil {
			return nil, err
		}
		if is {
			resource = v
		}
	}
	return resource, nil
}

type Resource interface {

	// Is 资源是否符合
	Is(url string, title string, w *selenium.WebDriver) (bool, error)

	// Name 资源名称
	Name() string

	// Before 爬虫前置操作,例如有些网站需要验证操作
	Before(w *selenium.WebDriver) error

	// Deal 整理资源地址,例如有些网站需要加上前缀等,或者过滤下载资源(同一个资源有多个分辨率的地址)
	Deal(w *selenium.WebDriver, urls []string) ([]string, error)
}

var Resources = []Resource{
	&_pornhub{},
	&_91porn{},
	&_51cg{},
}

type base struct{}

func (this *base) Is(url string, title string, w *selenium.WebDriver) (bool, error) {
	return true, nil
}

func (this *base) Name() string { return "base" }

func (this *base) Before(w *selenium.WebDriver) error { return nil }

func (thiss *base) Deal(w *selenium.WebDriver, urls []string) ([]string, error) {
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
	return urls, nil
}

/*

 */

type _pornhub struct{}

func (this *_pornhub) Is(url string, title string, w *selenium.WebDriver) (bool, error) {
	return strings.Contains(url, "pornhub.com"), nil
}

func (this *_pornhub) Name() string {
	return "Pornhub"
}

func (this *_pornhub) Before(w *selenium.WebDriver) error {
	//判断是否是已满18岁界面,如果是则点击确认
	e, err := w.FindElement(spider.ByXPATH, `//*[@id="modalWrapMTubes"]/div/div/button`)
	if err == nil {
		if err = e.Click(); err != nil {
			return err
		}
		<-time.After(time.Second * 3)
	}
	return nil
}

func (this *_pornhub) Deal(w *selenium.WebDriver, urls []string) ([]string, error) {
	list := []string{
		"1080P_",
		"720P_",
		"480P_",
		"240P_",
	}
	for _, v := range urls {
		logs.Trace("处理前资源: ", v)
	}
	for _, v := range list {
		for _, u := range urls {
			if strings.Contains(u, v) {
				return []string{u}, nil
			}
		}
	}
	return nil, nil
}

/*



 */

type _51cg struct{}

func (this *_51cg) Is(url string, title string, w *selenium.WebDriver) (bool, error) {
	return strings.Contains(title, "51cg"), nil
}

func (this *_51cg) Name() string {
	return "51cg"
}

func (this *_51cg) Before(w *selenium.WebDriver) error {
	return nil
}

func (this *_51cg) Deal(w *selenium.WebDriver, urls []string) ([]string, error) {
	for idx, v := range urls {
		urls[idx] = v + "&v=3&time=0"
	}
	return urls, nil
}

/*



 */

type _91porn struct{}

func (this *_91porn) Is(url string, title string, w *selenium.WebDriver) (bool, error) {
	return strings.Contains(url, "91porn"), nil
}

func (this *_91porn) Name() string {
	return "91porn"
}

func (this *_91porn) Before(w *selenium.WebDriver) error {
	return nil
}

func (this *_91porn) Deal(w *selenium.WebDriver, urls []string) ([]string, error) {
	//特殊处理91pron
	e, err := w.FindElement(selenium.ByXPATH, `//*[@id="player_one_html5_api"]/source`)
	if err != nil {
		return nil, err
	}
	url, err := e.GetAttribute("src")
	if err != nil {
		return nil, err
	}
	return []string{url}, nil
}
