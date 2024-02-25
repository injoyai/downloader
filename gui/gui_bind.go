package gui

//import (
//	"context"
//	"fmt"
//	"github.com/injoyai/selenium"
//	"github.com/injoyai/selenium/chrome"
//)
//
//func bindOpenBrowser(ctx context.Context, driverPath, browserPath string) error {
//	//如果seleniumServer没有启动，就启动一个seleniumServer所需要的参数，可以为空，示例请参见https://github.com/tebeka/selenium/blob/master/example_test.go
//	opts := []selenium.ServiceOption{}
//	selenium.SetDebug(true)
//	service, err := selenium.NewChromeDriverService(driverPath, 20165, opts...)
//	if nil != err {
//		return err
//	}
//	//注意这里，server关闭之后，chrome窗口也会关闭
//	defer service.Stop()
//
//	//链接本地的浏览器 chrome
//	caps := selenium.Capabilities{
//		"browserName": "chrome",
//	}
//	//禁止图片加载，加快渲染速度
//	pref := map[string]interface{}{}
//	// 模拟user-agent，防反爬
//	arg := []string{"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}
//	//设置浏览器参数
//	caps.AddChrome(chrome.Capabilities{
//		Path:  browserPath,
//		Prefs: pref,
//		Args:  arg,
//	})
//	// 调起chrome浏览器
//	web, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 20165))
//	if err != nil {
//		return err
//	}
//	_ = web
//	//web.Get("https://www.baidu.com")
//	<-ctx.Done()
//	return nil
//}
//
//func bindSetting() {
//
//}
//
//func bindDownload() {
//
//}
