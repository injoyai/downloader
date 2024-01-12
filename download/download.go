package download

import (
	"context"
	"github.com/injoyai/goutil/g"
)

func New() *Download {
	return NewWithContext(g.Ctx())
}

func NewWithContext(ctx context.Context) *Download {
	ctx, cancel := g.WithCancel(ctx)
	d := &Download{
		C:      make(chan *Task),
		ctx:    ctx,
		cancel: cancel,
	}
	go d.run()
	return d
}

type Download struct {
	C      chan *Task
	ctx    context.Context //ctx
	cancel context.CancelFunc
}

func (this *Download) Wait() {

}

func (this *Download) Append(t *Task) *Download {
	select {
	case <-this.ctx.Done():
		return this
	case this.C <- t:
	}
	return this
}

func (this *Download) Close() error {
	if this.cancel != nil {
		this.cancel()
	}
	return nil
}

func (this *Download) run() {
	for i := 0; ; i++ {
		select {
		case <-this.ctx.Done():
			return
		case t := <-this.C:
			t.Download()
		}
	}
}

type Option struct {
	Limit uint
	Retry uint
}

func (this *Option) new() *Option {
	if this == nil {
		this = new(Option)
	}
	op := &Option{
		Limit: this.Limit,
		Retry: this.Limit,
	}
	if op.Limit <= 0 {
		op.Limit = 20
	}
	if op.Retry <= 0 {
		op.Retry = 20
	}
	return op
}
