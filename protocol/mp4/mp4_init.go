package mp4

import (
	"context"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/task"
)

func NewTask(url string) []*task.Download {
	t := task.NewDownload()
	t.Append(task.GetBytesFunc(func(ctx context.Context, f func(p *http.Plan)) ([]byte, error) {
		return http.GetBytesWithPlan(url, f)
	}))
	return []*task.Download{t}
}
