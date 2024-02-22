package spider

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/selenium"
	"time"
)

type Element struct {
	selenium.WebElement
}

func (this *Element) Wait(t time.Duration) *Element {
	<-time.After(t)
	return this
}

func (this *Element) WaitSec(n ...int) *Element {
	return this.Wait(time.Duration(conv.GetDefaultInt(1, n...)) * time.Second)
}

func (this *Element) WaitMin(n ...int) *Element {
	return this.Wait(time.Duration(conv.GetDefaultInt(1, n...)) * time.Minute)
}

func (this *Element) Write(s string) error {
	return this.WebElement.SendKeys(s)
}

// ScreenshotBytes 截图,返回图片字节
func (this *Element) ScreenshotBytes(scroll bool) ([]byte, error) {
	return this.WebElement.Screenshot(scroll)
}

// ScreenshotSave 截图,并保存
func (this *Element) ScreenshotSave(filename string, scroll bool) error {
	bs, err := this.WebElement.Screenshot(scroll)
	if err != nil {
		return err
	}
	return oss.New(filename, bs)
}
