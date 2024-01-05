package gui

import (
	"fmt"
	"github.com/injoyai/goutil/net/http"
)

type IPlan interface {
	Run(func(p *Plan)) error
}

type Download struct {
	Url string
}

func (this *Download) Run(f func(p *Plan)) ([]byte, error) {
	return http.GetByteWithPlan(this.Url, func(p *http.Plan) {
		f(&Plan{
			Current: p.Current,
			Total:   p.Total,
			Message: fmt.Sprintf("%.2f%%", float64(p.Current)/float64(p.Total)*100),
		})
	})
}

type Chromedriver struct {
}

type Plan struct {
	Current int64  //当前数据
	Total   int64  //总数据
	Message string //消息
}
