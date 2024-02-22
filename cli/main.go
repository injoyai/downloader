package main

import (
	"context"
	"fmt"
	"github.com/injoyai/downloader/logic"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:     "download",
		Short:   "下载工具",
		Example: "download https://example.com",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				args = []string{"http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"}
			}
			if len(args) == 0 {
				fmt.Println("无下载地址")
				return
			}
			ctx := context.Background()
			b := bar.NewWithContext(ctx, 100)
			logs.PrintErr(logic.Download(
				ctx,
				logic.DefaultConfig.Dir,
				args[0],
				func(i *logic.Info) *logic.Info {
					b.SetTotal(i.Total)
					go b.Run()
					b.Add(i.Current)
					return i
				}, func(ctx context.Context, resp *task.DownloadItemResp) {
					b.Add(1)
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
