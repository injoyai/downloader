package spider

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/selenium"
	"strings"
	"time"
)

type WebDriver struct {
	*selenium.RemoteWD
}

// Wait 等待
func (this *WebDriver) Wait(t time.Duration) *WebDriver {
	<-time.After(t)
	return this
}

// WaitSecond 等待n秒
func (this *WebDriver) WaitSecond(n ...int) *WebDriver {
	return this.Wait(time.Duration(conv.GetDefaultInt(1, n...)) * time.Second)
}

// WaitMinute 等待n分钟
func (this *WebDriver) WaitMinute(n ...int) *WebDriver {
	return this.Wait(time.Duration(conv.GetDefaultInt(1, n...)) * time.Minute)
}

// Text 返回页面数据
func (this *WebDriver) Text() (string, error) {
	return this.PageSource()
}

// Open 打开网页
func (this *WebDriver) Open(url string) error {
	return this.Get(url)
}

// ScreenshotBytes 截图,返回图片字节
func (this *WebDriver) ScreenshotBytes() ([]byte, error) {
	return this.RemoteWD.Screenshot()
}

// ScreenshotSave 截图,并保存
func (this *WebDriver) ScreenshotSave(filename string) error {
	bs, err := this.RemoteWD.Screenshot()
	if err != nil {
		return err
	}
	return oss.New(filename, bs)
}

// FindAll 查找所有元素
func (this *WebDriver) FindAll(by, value string) ([]*Element, error) {
	es, err := this.FindElements(by, value)
	if err != nil {
		return nil, err
	}
	list := []*Element(nil)
	for _, v := range es {
		list = append(list, &Element{v})
	}
	return list, nil
}

// Find 查找一个元素
func (this *WebDriver) Find(by, value string) (*Element, error) {
	e, err := this.FindElement(by, value)
	return &Element{e}, err
}

// FindTagAttributes 查找所有标签的属性Attribute,例如a.href
func (this *WebDriver) FindTagAttributes(tag string) ([]string, error) {
	tagList := strings.Split(tag, ".")
	name := tagList[len(tagList)-1]
	es, err := this.FindTags(strings.Join(tagList[:len(tagList)-1], "."))
	if err != nil {
		return nil, err
	}
	list := []string(nil)
	for _, v := range es {
		//标签不一定有这个属性,固忽略错误
		s, err := v.GetAttribute(name)
		if err == nil {
			switch s {
			case "", "javascript:;":
			default:
				list = append(list, s)
			}
		}
	}
	return list, nil
}

// RangeTags 遍历标签,例如a标签
func (this *WebDriver) RangeTags(tag string, f func(*Element) error) error {
	es, err := this.FindTags(tag)
	if err != nil {
		return err
	}
	for _, e := range es {
		if err = f(e); err != nil {
			return err
		}
	}
	return nil
}

// FindTags 查找所有标签,例如a标签,href在a标签里面
func (this *WebDriver) FindTags(tag string) ([]*Element, error) {
	var es []selenium.WebElement
	var err error
	tagList := strings.Split(tag, ".")
	for i, v := range tagList {
		if i == 0 {
			es, err = this.FindElements(ByTagName, v)
			if err != nil {
				return nil, err
			}
		} else {
			es2 := []selenium.WebElement(nil)
			for _, e := range es {
				es, err = e.FindElements(ByTagName, v)
				if err != nil {
					return nil, err
				}
				es2 = append(es2, es...)
			}
			es = es2
		}
	}
	list := []*Element(nil)
	for _, v := range es {
		list = append(list, &Element{v})
	}
	return list, nil
}

// FindTag 查找标签,例如a标签
func (this *WebDriver) FindTag(tag string) (*Element, error) {
	e, err := this.FindElement(ByTagName, tag)
	return &Element{e}, err
}

// FindXPaths 查找所有XPath
func (this *WebDriver) FindXPaths(path string) ([]*Element, error) {
	es, err := this.FindElements(ByXPATH, path)
	if err != nil {
		return nil, err
	}
	list := []*Element(nil)
	for _, v := range es {
		list = append(list, &Element{v})
	}
	return list, nil
}

// FindXPath 查找所有XPath
func (this *WebDriver) FindXPath(path string) (*Element, error) {
	e, err := this.FindElement(ByXPATH, path)
	return &Element{e}, err
}

// FindSelects 查找所有Select
func (this *WebDriver) FindSelects(path string) ([]*Element, error) {
	es, err := this.FindElements(ByCSSSelector, path)
	if err != nil {
		return nil, err
	}
	list := []*Element(nil)
	for _, v := range es {
		list = append(list, &Element{v})
	}
	return list, nil
}

// FindSelect 查找所有Select
func (this *WebDriver) FindSelect(path string) (*Element, error) {
	e, err := this.FindElement(ByCSSSelector, path)
	return &Element{e}, err
}
