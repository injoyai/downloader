package main

import (
	"context"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/logic"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/win"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func main() {

	logs.SetWriter(logs.Stdout)

	root := &command.Command{
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
			{Name: "proxyEnable", Default: "true", Memo: "是否使用HTTP代理"},
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
	switch args[0] {
	case "registerUrlProtocol":
		registerUrlProtocol(cmd, args, flags)
		return

	case "test":
		args[0] = "http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"
	}

	cfg := &logic.Config{
		Source:       args[0],
		Dir:          flags.GetString("dir"),
		Name:         flags.GetString("name"),
		Retry:        flags.GetUint("retry"),
		Coroutine:    flags.GetUint("coroutine"),
		ProxyEnable:  flags.GetBool("proxyEnable"),
		ProxyAddress: flags.GetString("proxyAddress"),
		NoticeEnable: flags.GetBool("noticeEnable"),
		NoticeText:   flags.GetString("noticeText"),
		VoiceEnable:  flags.GetBool("voiceEnable"),
		VoiceText:    flags.GetString("voiceText"),
	}

	err := logic.Download(context.Background(), cfg)
	fmt.Println("下载结果: ", conv.New(err).String(cfg.Filename()))
}

func registerUrlProtocol(cmd *cobra.Command, args []string, flags *command.Flags) {
	err := win.RegisterURLProtocol(win.REGISTER_ROOT, "download", oss.ExecName())
	fmt.Println("注册结果: ", conv.New(err).String("成功"))
}
