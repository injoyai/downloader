package main

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
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
			start := time.Now()
			cache := maps.NewValue(0, time.Millisecond)
			b := bar.NewWithContext(ctx, 100)
			b.SetFormatter(func(e *bar.Format) string {
				return fmt.Sprintf("\r%s  %s  %s",
					e.Bar,
					e.Size,
					func() string {
						if v, ok := cache.Val(); ok {
							return v.(string)
						}

						f, unit := oss.Size(int64(float64(length) / time.Now().Sub(start).Seconds()))
						if f < 0 {
							f, unit = 0, "B"
						}
						s := fmt.Sprintf("%0.1f%s/s", f, unit)
						if f > 0 {
							cache = maps.NewValue(s, time.Millisecond*500)
						}
						return s
					}(),
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

	//root.AddCommand(
	//	&cobra.Command{
	//		Use: "config",
	//		Run: func(cmd *cobra.Command, args []string) {
	//			fmt.Println(logic.DefaultConfig)
	//		},
	//	},
	//	&cobra.Command{
	//		Use: "set",
	//		Run: func(cmd *cobra.Command, args []string) {
	//			fmt.Println(logic.DefaultConfig)
	//		},
	//	},
	//)

	logs.PrintErr(root.Execute())
}
