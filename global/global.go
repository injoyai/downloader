package global

import (
	_ "embed"
	"github.com/injoyai/goutil/oss"
)

var (
	//go:embed chrome.zip
	ChromeZip []byte

	// BrowserDir 浏览器目录
	BrowserDir = oss.UserInjoyDir("downloader/browser")

	// ChromePath 浏览器路径
	ChromePath = oss.UserInjoyDir("/browser/chrome/chrome.exe")

	// DriverPath 浏览器驱动路径
	DriverPath = oss.UserInjoyDir("/browser/chrome/chromedriver.exe")

	// ConfigPath 配置目录
	ConfigPath = oss.UserInjoyDir("downloader/config/config.json")
)
