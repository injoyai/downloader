package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/spider"
	"github.com/injoyai/logs"
	"strings"
)

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
		logs.Debug(w.Status())
		logs.Debug("-------------------------------------------")
		list, err := w.FindTagAttributes("a.href")
		g.PanicErr(err)
		logs.Debug(strings.Join(list, "\n"))
		bs, _ := w.Screenshot()
		oss.New("./build.png", bs)
		w.WaitSecond(2)
		return nil
	}))
}
