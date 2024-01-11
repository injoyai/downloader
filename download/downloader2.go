package download

import (
	"context"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io"
)

type Download struct {
	C   chan *ITask
	ctx context.Context //ctx
}

func (this *Download) Append(t *ITask) {
	select {
	case <-this.ctx.Done():
		return
	case this.C <- t:
	}
}

func (this *Download) Run() {
	for i := 0; ; i++ {
		select {
		case <-this.ctx.Done():
			return
		case t := <-this.C:
			t.Download()
		}
	}
}

type ITask struct {
	Queue    chan GetBytes          //分片队列
	limit    uint                   //协程数
	retry    uint                   //重试次数
	offset   int                    //偏移量
	writer   io.Writer              //
	doneItem func(i int, err error) //
	doneAll  func()                 //
}

func (this *ITask) Download() {
	wg := chans.NewWaitLimit(this.limit)
	cache := make([][]byte, len(this.Queue))
	for i := 0; i < len(this.Queue); i++ {
		select {
		case v := <-this.Queue:
			if cache[i] == nil {
				wg.Add()
				go func(i int, t *ITask) {
					defer wg.Done()
					bytes, err := t.GetBytes(v)
					if err == nil {
						cache[i] = bytes
					}
					this.doneItem(i, err)
				}(i, this)
			}
		}
	}
	wg.Wait()
	for _, bs := range cache {
		this.writer.Write(bs)
	}
}

func (this *ITask) RetryNum() int {
	if this.retry <= 0 {
		return 1
	}
	return int(this.retry)
}

func (this *ITask) GetBytes(v GetBytes) (bytes []byte, err error) {
	for i := 0; i < this.RetryNum(); i++ {
		bytes, err = v.GetBytes()
		if err == nil {
			return
		}
	}
	return
}

type GetBytes interface {
	GetBytes() ([]byte, error)
}
