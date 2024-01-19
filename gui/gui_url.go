package gui

import (
	"errors"
	"fmt"
	"github.com/injoyai/downloader/protocol/m3u8"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/logs"
	"github.com/tebeka/selenium"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	Regexp = regexp.MustCompile(`(http://|https://)[a-zA-Z0-9\\/=_\-.:%&]+\.(m3u8|mp4)([\?a-zA-Z0-9/=_\-.*%&]+)`)
)

func RegexpAll(s string) []string {
	return Regexp.FindAllString(s, -1)
}

func (this *Config) deepFind(p spider.Page) ([]string, error) {
	urls := m3u8.RegexpAll(p.String())
	//urls := RegexpAll(p.String())
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

// FindUrlWithSelenium 通过资源地址获取到下载连接
func (this *Config) FindUrlWithSelenium(driverPath, browserPath string) (urls []string, err error) {

	u := this.DownloadAddr

	if strings.Contains(u, ".m3u8") {
		return []string{u}, nil
	}

	if !strings.Contains(u, "http") {
		return nil, errors.New("无效资源地址")
	}

	logs.Debug("驱动位置: ", driverPath)
	logs.Debug("浏览器位置: ", browserPath)
	logs.Debug("开始爬取: ", u)
	if err := spider.New(driverPath, browserPath).
		ShowWindow(false).ShowImg(false).Run(func(i spider.IPage) {
		p := i.Open(u)
		p.WaitSec(3)

		title, _ := p.Title()

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

		case strings.Contains(title, "51cg"):

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

	logs.Debug("爬取成功: ", urls)

	if len(urls) == 0 {
		return nil, errors.New("没有找到资源")
	}

	return urls, nil
}