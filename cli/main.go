package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/injoyai/downloader/logic"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"time"
)

func main() {
	root := &cobra.Command{
		Use:     "download",
		Short:   "下载工具",
		Example: "download https://example.com",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				//测试
				args = []string{"http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"}
			}
			if len(args) == 0 {
				fmt.Println("无下载地址")
				return
			}
			ctx := context.Background()
			length := int64(0)
			b := bar.NewWithContext(ctx, 100)
			b.SetColor(color.BgCyan)
			b.SetFormatter(func(e *bar.Format) string {
				f1, unit1 := oss.Size(length)
				return fmt.Sprintf("\r%s  %s  %s  %s",
					e.Bar,
					e.Size,
					fmt.Sprintf("%0.1f%s", f1, unit1),
					b.Speed("speed", length, time.Millisecond*500),
				)
			})
			logs.PrintErr(logic.Download(
				ctx,
				args[0],
				func(i *logic.Info) *logic.Info {
					b.SetTotal(i.Total)
					go b.Run()
					b.Add(i.Current)
					return i
				}, func(ctx context.Context, resp *task.DownloadItemResp) {
					b.Add(1)
					length += int64(len(resp.Bytes))
				},
			))
		},
	}

	logs.PrintErr(root.Execute())
}
