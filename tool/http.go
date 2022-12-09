package tool

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var HTTP = client()

func client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			//连接结束后会直接关闭,
			//否则会加到连接池复用
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				//设置可以访问HTTPS
				InsecureSkipVerify: true,
			},
		},
		//设置连接超时时间,连接成功后无效
		//连接成功后数据读取时间可以超过这个时间
		//数据读取超时等可以nginx配置
		Timeout: time.Second * 10,
	}
}

func ProxyClient(proxyUrl string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			//连接结束后会直接关闭,
			//否则会加到连接池复用
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				//设置可以访问HTTPS
				InsecureSkipVerify: true,
			},
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(proxyUrl)
			},
		},
		//设置连接超时时间,连接成功后无效
		//连接成功后数据读取时间可以超过这个时间
		//数据读取超时等可以nginx配置
		Timeout: time.Second * 10,
	}
}

func GetBytes(url string) ([]byte, error) {
	resp, err := HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
