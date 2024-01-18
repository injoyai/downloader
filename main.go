package main

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/gui"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/logs"
)

//go:embed chrome.zip
var chromeZip []byte

var Debug = "true"

func init() {
	logs.SetLevel(conv.Select(Debug == "true", logs.LevelAll, logs.LevelNone).(logs.Level))
}

// http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8
// https://www.wangfei.tv/vodplay/302601-3-1.html
func main() {
	if err := spider.Install(chromeZip); err != nil {
		fmt.Println("爬虫工具安装失败:", err.Error())
	}
	logs.PrintErr(gui.New())
}
