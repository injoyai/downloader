package tool

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"sync"
)

// Speak 生成音频并播放
func Speak(msg string) error {
	return NewVoice().Speak(msg)
}

func NewVoice() *Voice {
	return &Voice{
		Rate:   0,
		Volume: 100,
	}
}

var mu sync.Mutex

type Voice struct {
	Rate   int //语速
	Volume int //音量
}

func (this *Voice) SetRate(n int) *Voice {
	this.Rate = n
	return this
}

func (this *Voice) SetVolume(n int) *Voice {
	if n > 100 {
		n = 100
	} else if n < 0 {
		n = 0
	}
	this.Volume = n
	return this
}

func (this *Voice) Speak(msg string) (err error) {
	mu.Lock()
	defer mu.Unlock()
	defer Recover(&err)
	PanicErr(ole.CoInitialize(0))
	unknown, err := oleutil.CreateObject("SAPI.SpVoice")
	PanicErr(err)
	voice, err := unknown.QueryInterface(ole.IID_IDispatch)
	PanicErr(err)
	_, err = oleutil.PutProperty(voice, "Rate", this.Rate)
	PanicErr(err)
	_, err = oleutil.PutProperty(voice, "Volume", this.Volume)
	PanicErr(err)
	_, err = oleutil.CallMethod(voice, "Speak", msg)
	PanicErr(err)
	_, err = oleutil.CallMethod(voice, "WaitUntilDone", 0)
	PanicErr(err)
	voice.Release()
	ole.CoUninitialize()
	return nil
}
