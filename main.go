package main

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/downloader/global"
	"github.com/injoyai/downloader/gui"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/logs"
)

var Debug = "true"

func init() {
	logs.SetLevel(conv.Select(Debug == "true", logs.LevelAll, logs.LevelNone).(logs.Level))
}

// http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8
// https://www.wangfei.tv/vodplay/302601-3-1.html
func main() {
	logs.PrintErr(spider.Install(
		global.BrowserDir,
		global.ChromeZip,
	))
	logs.PrintErr(gui.New(
		global.ConfigPath,
		global.DriverPath,
		global.ChromePath,
	))
}
