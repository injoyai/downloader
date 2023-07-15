package main

import (
	"fmt"
	"github.com/injoyai/downloader/gui"
	"github.com/injoyai/downloader/spider"
	"github.com/injoyai/logs"
)

// http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8
func main() {
	if err := spider.Install(); err != nil {
		fmt.Println("爬虫工具安装失败:", err.Error())
	}
	logs.PrintErr(gui.New())
}
