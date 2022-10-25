package download

import (
	"context"
	"github.com/injoyai/downloader/tool"
	"io"
	"sync"
	"time"
)

func New(op *Option) *Downloader {
	return NewWithContext(context.Background(), op)
}

func NewWithContext(ctx context.Context, op *Option) *Downloader {
	op = op.new()
	ctx, cancel := context.WithCancel(ctx)
	return &Downloader{
		limit:  op.Limit,
		bar:    tool.NewBar().SetColor(op.BarColor).SetStyle(op.BarStyle),
		ctx:    ctx,
		cancel: cancel,
		retry:  op.Retry,
	}
}

type Downloader struct {
	queue  chan Item          //普通队列
	limit  uint               //协程数
	bar    *tool.Bar          //进度条
	ctx    context.Context    //ctx
	cancel context.CancelFunc //cancel
	retry  uint               //重试次数
	err    []error            //错误
}

func (this *Downloader) addErr(err error) {
	if err != nil {
		this.err = append(this.err, err)
	}
}

func (this *Downloader) Bar() *tool.Bar {
	return this.bar
}

func (this *Downloader) Retry() int {
	if this.retry <= 0 {
		return 1
	}
	return int(this.retry)
}

func (this *Downloader) runTask(t Item) (bytes []byte, err error) {
	for i := 0; i < this.Retry(); i++ {
		bytes, err = t.Run()
		if err == nil {
			return
		}
	}
	return
}

func (this *Downloader) Run(list Task, writer io.Writer) []error {
	this.bar.SetSize(float64(list.Len()))
	this.queue = make(chan Item, list.Len())
	cache := make([][]byte, list.Len()+1)
	for _, v := range list.List() {
		this.queue <- v
	}
	idx := 0
	wg := sync.WaitGroup{}
	ch := make(chan byte, this.limit)
	fn := func(ctx context.Context, c chan Item) {
		for {
			select {
			case <-ctx.Done():
				return
			case i := <-c:
				wg.Add(1)
				ch <- 0
				if cache[i.Idx()] == nil {
					go func(t Item) {
						defer func() {
							this.bar.Add(1)
							<-ch
							wg.Done()
						}()
						bytes, err := this.runTask(t)
						this.addErr(err)
						if err == nil {
							cache[i.Idx()] = bytes
							for {
								if bs := cache[idx]; bs != nil {
									writer.Write(bs)
									idx++
									continue
								}
								break
							}
						}
					}(i)
				}
			}
		}
	}
	go fn(this.ctx, this.queue)
	go this.bar.Wait()
	time.Sleep(time.Second)
	wg.Wait()
	return this.err
}

type Option struct {
	Limit    uint
	Retry    uint
	BarColor int
	BarStyle string
}

func (this *Option) new() *Option {
	if this == nil {
		this = new(Option)
	}
	op := &Option{
		Limit:    this.Limit,
		Retry:    this.Limit,
		BarColor: this.BarColor,
		BarStyle: this.BarStyle,
	}
	if op.Limit <= 0 {
		op.Limit = 20
	}
	if op.Retry <= 0 {
		op.Retry = 20
	}
	if len(op.BarStyle) == 0 {
		op.BarStyle = ">"
	}
	return op
}
