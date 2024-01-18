package global

import (
	_ "embed"
	"github.com/injoyai/goutil/oss"
	"path/filepath"
)

var (
	//go:embed chrome.zip
	ChromeZip []byte

	// BrowserDir 浏览器目录
	BrowserDir = oss.UserInjoyDir("downloader/browser")

	// ChromePath 浏览器路径
	ChromePath = filepath.Join(BrowserDir, "chrome/chrome.exe")

	// DriverPath 浏览器驱动路径
	DriverPath = filepath.Join(BrowserDir, "chrome/chromedriver.exe")

	// ConfigPath 配置目录
	ConfigPath = oss.UserInjoyDir("downloader/config/config.json")
)
