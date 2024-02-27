package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/injoyai/downloader/logic"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"time"
)

func main() {

	root := command.Command{
		Command: cobra.Command{
			Use:     "download",
			Short:   "下载工具",
			Example: "download https://example.com",
		},
		Flag: []*command.Flag{
			{Name: "name", Short: "n", Memo: "自定义保存名称"},
			{Name: "retry", Default: "10", Memo: "失败重试次数"},
			{Name: "coroutine", Short: "c", Default: "20", Memo: "协程数量"},
			{Name: "dir", Default: "./", Memo: "协程数量"},
			{Name: "suffix", Default: ".ts", Memo: "文件后缀"},
			{Name: "proxyEnable", Default: "false", Memo: "是否使用HTTP代理"},
			{Name: "proxyAddress", Default: "http://127.0.0.1:1081", Memo: "HTTP代理地址"},
			{Name: "noticeEnable", Default: "true", Memo: "是否启用通知"},
			{Name: "noticeText", Default: "主人. 您的视频已下载结束", Memo: "通知内容"},
			{Name: "voiceEnable", Default: "true", Memo: "是否启用语音"},
			{Name: "voiceText", Default: "主人. 您的视频已下载结束", Memo: "语音内容"},
		},
		Run: handler,
	}

	logs.PrintErr(root.Execute())
}

func handler(cmd *cobra.Command, args []string, flags *command.Flags) {
	if len(args) == 0 {
		fmt.Println("无下载地址")
		return
	}
	if args[0] == "test" {
		//测试
		args = []string{"http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"}
	}

	ctx := context.Background()
	sum := int64(0)
	current := int64(0)
	b := bar.NewWithContext(ctx, 0)
	b.SetColor(color.BgCyan)
	b.SetFormatter(func(e *bar.Format) string {
		return fmt.Sprintf("\r%s  %s  %s  %s",
			e.Bar,
			e.RateSize,
			oss.SizeString(sum),
			b.SpeedUnit("speed", current, time.Millisecond*500),
		)
	})

	logs.PrintErr(logic.Download(
		ctx,
		args[0],
		func(i *logic.Info) *logic.Info {
			b.SetTotal(i.Total)
			go b.Run()
			i.Name = flags.GetString("name")
			i.Retry = flags.GetUint("retry")
			i.Coroutine = flags.GetUint("coroutine")
			i.Dir = flags.GetString("dir")
			i.Suffix = flags.GetString("suffix")
			i.ProxyEnable = flags.GetBool("proxyEnable")
			i.ProxyAddress = flags.GetString("proxyAddress")
			i.NoticeEnable = flags.GetBool("noticeEnable")
			i.NoticeText = flags.GetString("noticeText")
			i.VoiceEnable = flags.GetBool("voiceEnable")
			i.VoiceText = flags.GetString("voiceText")
			return i
		}, func(ctx context.Context, resp *task.DownloadItemResp) {
			current = resp.GetSize()
			sum += current
			b.Add(1)
		},
	))

}
