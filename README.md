## m3u8资源下载工具

## 工具介绍

* 需要Google浏览器(Chrome),图形化界面(GUI)和爬虫功能依赖此流浪器
* chromedriver驱动版本目前(2023-12)只支持v114.xxxx.xx及以下版本的浏览器,否则爬虫功能不可用
* Chrome历史版本下载地址(https://chromedriver.storage.googleapis.com/index.html)
* 提供图形界面,使用lorca(https://github.com/zserge/lorca)实现,需要Google浏览器(Chrome)
* 目前只支持m3u8资源下载,后续增加其它类型的资源
* 如果安装了Chrome,工具会下载chromedriver.exe到根目录,提供爬虫功能,能爬取普通网页(支持js,动态加载,iframe)的所有m3u8资源
* 下载完成后,默认后缀为.ts
* 显示下载进度,下载用时



## 使用说明
* [版本下载.windows](https://github.com/injoyai/downloader/releases/latest/download/downloader.exe)
* 下载地址(例 http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8)
  或者普通网页地址(例 https://www.wangfei.tv/vodplay/302601-3-1.html)
* 保存名称可选,重命名文件(可选,xx,xx.ts,xx.mp4) ,存在相同名字文件会被覆盖
* 等待进度条完成,或显示下载成功xxx ,则完成下载
  ![](doc/downloader.png)

## 测试结果

  |网站|m3u8|html| 说明                                                              | 测试时间       |
  |---|---|---|-----------------------------------------------------------------|------------|
  |任意.m3u8资源|√|-| 后缀是.m3u8即可                                                      |            |
  |https://www.acfun.cn|√|√| 一个页面有多个资源,不同清晰度,会全部下载                                           |            |
  |https://www.wangfei.tv|√|√| 中规中矩                                                            |            |
  |https://www.91porn.com|√|√| 有一天15次限制,正则VID=[0-9]+,得到https://cdn77.91p49.com/m3u8/%s/%s.m3u8 |            |
  |https://zxzj.vip|√|√| 网页有iframe嵌套                                                     |            |
  |https://jable.tv|√|√|                                                                 | 2023-12-28 |
  |https://51cg.fun|√|√|                                                                 | 2023-12-29 |

## 更新说明
有人点赞的话,就更新一波
1.

## 常见问题
1. 开始能正常使用的爬虫功能,一段时间后却不能使用了,可能原因是浏览器升级了,驱动版本不兼容,解决方法删除驱动文件chromedriver.exe,并重新打开工具
2. 