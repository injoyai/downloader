package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/injoyai/downloader/download"
	"github.com/injoyai/downloader/download/m3u8"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/downloader/tool"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Download() {
	application := app.New()
	window := application.NewWindow("Downloader")
	window.SetContent(makeMainTab())
	window.Resize(fyne.NewSize(600, 300))
	window.ShowAndRun()
}

func makeMainTab() *fyne.Container {
	_input := NewInput("address")
	_inputDir := NewInput(tool.Cfg.Dir())
	_inputFilename := NewInput("")
	_scroll := NewScroll().SetSize(600, 200)
	_button := NewButton("Start").SetOnclick(func(b *Button) {
		_time := time.Now()
		err := onclick(b, _scroll, _input.Text, _inputDir.Text, _inputFilename.Text)
		if err != nil {
			_scroll.SetText(err.Error())
			return
		}
		_scroll.SetText("Download success  Used " + time.Now().Sub(_time).String())
	})
	_button.Resize(fyne.NewSize(600, 200))
	return container.NewVBox(
		NewLabel("download url"),
		_input,
		NewLabel("download dir"),
		_inputDir,
		NewLabel("download name"),
		_inputFilename,
		_button,
		_scroll,
	)
}

func findUrl(u string) ([]string, error) {
	urls := []string(nil)
	if strings.Contains(u, ".m3u8") {
		return []string{u}, nil
	}

	if !strings.Contains(u, "http") {
		return nil, errors.New("invalid url")
	}
	spider.New("./chromedriver.exe").ShowWindow(false).ShowImg(false).Run(func(i spider.IPage) {
		p := i.Open(u)
		p.WaitSec(3)
		for x := 0; x < 5; x++ {
			urls = regexp.MustCompile(`(http://|https://)[a-zAA-Z0-9/\-.]+\.m3u8`).FindAllString(p.String(), -1)
			if len(urls) > 0 {
				break
			}
			p.WaitSec()
		}
	})

	for _, v := range urls {
		if strings.Contains(v, `//test.`) {
			host := tool.CropLast(v, "/")
			bs, _ := tool.GetBytes(host)
			s := regexp.MustCompile(`>(.*?)\.m3u8<`).FindString(string(bs))
			s = tool.CropFirst(s, ">", false)
			s = tool.CropLast(s, "<", false)
			v = host + s
		}
	}

	return urls, nil
}

func onclick(b *Button, s *Scroll, text, downloadDir, filename string) (err error) {
	if len(downloadDir) == 0 {
		downloadDir = tool.Cfg.Dir()
	}

	if tool.Cfg.Dir() != downloadDir {
		tool.Cfg.DownloadDir = downloadDir
		tool.Cfg.Save()
	}

	if err := os.MkdirAll(downloadDir, 0777); err != nil {
		return err
	}

	b.SetText("Stop")

	if len(text) == 0 {
		return errors.New("invalid url")
	}

	s.SetText(text)
	urls, err := findUrl(text)
	if err != nil {
		return err
	}
	list := make([]string, len(urls))
	wg := sync.WaitGroup{}
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			list[i] = url
			s.SetText(strings.Join(list, "\n"))
			l, err := m3u8.NewTask(url)
			if err != nil {
				list[i] = err.Error()
				return
			}
			if len(filename) == 0 {
				filename = l.Filename()
			} else if !strings.Contains(filename, ".") {
				filename += filepath.Ext(l.Filename())
			}
			f, err := os.Create(downloadDir + filename)
			if err != nil {
				list[i] = err.Error()
				return
			}

			d := download.New(nil)

			d.Bar().SetPrefix("plan:").
				SetPrint(func(x string) {
					list[i] = x
					s.SetText(strings.Join(list, "\n"))
				})

			errs := d.Run(l, f)
			f.Close()
			if len(errs) > 0 {
				list[i] = errs[0].Error()
			} else {
				list[i] = "Success"
			}
			s.SetText(strings.Join(list, "\n"))
		}(i, url)
	}
	wg.Wait()
	return nil
}

//=========================Label=======================

func NewLabel(text string) *widget.Label {
	return widget.NewLabel(text)
}

//=========================Scroll=======================

func NewScroll() *Scroll {
	e := widget.NewMultiLineEntry()
	e.Disable()
	slo := container.NewScroll(e)
	return &Scroll{
		e:      e,
		Scroll: slo,
		//c:      cache.NewCycle(100),
	}
}

type Scroll struct {
	e *widget.Entry
	*container.Scroll
	//c    *cache.Cycle
	w, h float32
}

func (this *Scroll) SetSize(w, h float32) *Scroll {
	this.w, this.h = w, h
	this.e.Resize(fyne.NewSize(w, h))
	this.Scroll.Resize(fyne.NewSize(w, h))
	return this
}

func (this *Scroll) SetText(s string) *Scroll {
	this.e.SetText(s)
	this.SetSize(this.w, this.h)
	return this
}

//func (this *Scroll) AddText(s string) *Scroll {
//	this.c.Add(s)
//	list := []string{}
//	for _, v := range this.c.List(20) {
//		list = append(list, v.String())
//	}
//	this.SetText(strings.Join(list, "\n"))
//	return this
//}

//=========================Radio=======================

func NewRadio(list []string) *Radio {
	r := &Radio{}
	r.RadioGroup = widget.NewRadioGroup(list, r._changed)
	r.RadioGroup.Horizontal = true
	return r
}

type Radio struct {
	*widget.RadioGroup
	changed func(string)
}

func (this *Radio) _changed(s string) {
	if this.changed != nil {
		this.changed(s)
	}
}

func (this *Radio) Changed(fn func(string)) *Radio {
	this.changed = fn
	return this
}

//=========================Input=======================

func NewInput(hint string, label ...string) *widget.Entry {
	input := widget.NewEntry()
	input.SetPlaceHolder(hint)
	if len(label) > 0 {
		input.SetText(label[0])
	}
	return input
}

//=========================Button=======================

func NewButton(label string) *Button {
	b := &Button{}
	b.Button = widget.NewButton(label, b._onclick)
	go func() {
		for i := 0; ; i++ {
			time.Sleep(time.Second)
			b.SetText(strconv.Itoa(i))
		}
	}()
	return b
}

type Button struct {
	*widget.Button
	label   string
	onclick func(*Button)
}

func (this *Button) _onclick() {
	if this.onclick != nil {
		this.onclick(this)
	}
}

func (this *Button) SetOnclick(fn func(*Button)) *Button {
	this.onclick = fn
	return this
}
