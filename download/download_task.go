package download

import (
	"context"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/goutil/g"
	"io"
)

func NewTask() *Task {
	return NewTaskWithContext(g.Ctx())
}

func NewTaskWithContext(ctx context.Context) *Task {
	ctx, cancel := g.WithCancel(ctx)
	return &Task{
		ctx:    ctx,
		cancel: cancel,
	}
}

type Task struct {
	queue    []GetBytes             //分片队列
	limit    uint                   //协程数
	retry    uint                   //重试次数
	offset   int                    //偏移量
	writer   io.Writer              //
	doneItem func(i int, err error) //
	doneAll  func()                 //
	ctx      context.Context        //
	cancel   context.CancelFunc     //
}

func (this *Task) Len() int {
	return len(this.queue)
}

func (this *Task) Append(v GetBytes) *Task {
	this.queue = append(this.queue, v)
	return this
}

func (this *Task) SetLimit(limit uint) *Task {
	this.limit = limit
	return this
}

func (this *Task) SetRetry(retry uint) *Task {
	this.retry = retry
	return this
}

func (this *Task) SetWriter(writer io.Writer) *Task {
	this.writer = writer
	return this
}

func (this *Task) SetDoneItem(doneItem func(i int, err error)) *Task {
	this.doneItem = doneItem
	return this
}

func (this *Task) SetDoneAll(doneAll func()) *Task {
	this.doneAll = doneAll
	return this
}

func (this *Task) Download() {
	wg := chans.NewWaitLimit(this.limit)
	cache := make([][]byte, this.Len())
	for i, v := range this.queue {
		wg.Add()
		go func(i int, t *Task, v GetBytes) {
			defer wg.Done()
			bytes, err := t.getBytes(v)
			if err == nil {
				cache[i] = bytes
			}
			if this.doneItem != nil {
				this.doneItem(i, err)
			}
		}(i, this, v)
	}
	wg.Wait()
	for _, bs := range cache {
		this.writer.Write(bs)
	}
	if this.doneAll != nil {
		this.doneAll()
	}
}

func (this *Task) retryNum() int {
	if this.retry <= 0 {
		return 1
	}
	return int(this.retry)
}

func (this *Task) getBytes(v GetBytes) (bytes []byte, err error) {
	for i := 0; i < this.retryNum(); i++ {
		bytes, err = v.GetBytes(this.ctx)
		if err == nil {
			return
		}
	}
	return
}

type GetBytes interface {
	GetBytes(ctx context.Context) ([]byte, error)
}
