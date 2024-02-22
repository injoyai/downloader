package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/spider"
	"github.com/injoyai/logs"
	"strings"
)

/*
获取所有a标签的href
*/
func main() {
	logs.PrintErr(spider.New(
		oss.UserInjoyDir("/browser/chrome/chrome.exe"),
		oss.UserInjoyDir("/browser/chrome/chromedriver.exe"),
		func(e *spider.Entity) {
			e.ShowWindow(false)
			e.ShowImg(false)
		},
	).Run(func(w *spider.WebDriver) error {
		w.Open("http://www.baidu.com")
		list, err := w.FindTagAttributes("a.href")
		g.PanicErr(err)
		logs.Debug(strings.Join(list, "\n"))
		w.WaitSecond(2)
		return nil
	}))
}
