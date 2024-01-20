package m3u8

import (
	"context"
	"encoding/hex"
	"github.com/injoyai/base/bytes/crypt/aes"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/goutil/task"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

func RegexpAll(s string) []string {
	return regexp.MustCompile(`(http)[a-zA-Z0-9\\/=_\-.:%&]+\.m3u8([?&a-zA-Z0-9/=_\-.]+)`).FindAllString(s, -1)
}

func NewResponse(uri string) (*Response, error) {
	bs, err := http.GetBytes(uri)
	if err != nil {
		return nil, err
	}
	//解析数据
	return Decode(uri, bs)
}

func Decode(uri string, bs []byte) (resp *Response, err error) {
	host, err := url.Parse(str.CropLast(uri, "/"))
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	resp = &Response{host: host, filename: filepath.Base(base.Path)}
	nextItem := false
	nextM3u8 := false
	for _, s := range strings.Split(string(bs), "\n") {
		switch true {
		case nextItem:
			if !strings.HasPrefix(s, "http") {
				//相对路径
				suffixURL, err := url.Parse(s)
				if err != nil {
					return nil, err
				}
				s = resp.host.ResolveReference(suffixURL).String()
			}
			resp.TS_URL = append(resp.TS_URL, &Url{
				url:    s,
				isM3u8: nextM3u8,
			})
			nextItem = false
			nextM3u8 = false
		case strings.HasPrefix(s, "#EXT-X-KEY:"):
			s = strings.TrimPrefix(s, "#EXT-X-KEY:")
			//按,分割
			for _, v := range strings.Split(s, ",") {
				if list := strings.SplitN(v, "=", 2); len(list) == 2 {
					switch list[0] {
					case "METHOD":
						//加密方式
						resp.Method = list[1]
					case "URI":
						//秘钥地址
						if !strings.HasPrefix(s, "http") {
							suffixURL, err := url.Parse(strings.Trim(list[1], `"`))
							if err != nil {
								return nil, err
							}
							s = resp.host.ResolveReference(suffixURL).String()
						}
						resp.Key, err = http.GetBytes(s)
						if err != nil {
							return nil, err
						}
					case "IV":
						//秘钥
						if len(list[1]) > 2 && strings.ToLower(list[1][:2]) == "0x" {
							resp.IV, err = hex.DecodeString(list[1][2:])
							if err != nil {
								return nil, err
							}
						} else {
							//todo
						}
					}
				}
			}
		case strings.HasPrefix(s, "#EXTINF:") || strings.HasPrefix(s, "#EXT-X-STREAM-INF"):
			//下一行是下载地址
			nextItem = true
			nextM3u8 = strings.HasPrefix(s, "#EXT-X-STREAM-INF")

		case strings.HasPrefix(s, "#EXT-X-ENDLIST"):
			//列表结束
			break
		}
	}
	return
}

type Response struct {
	filename string   //文件名称
	host     *url.URL //主机,前缀
	TS_URL   []*Url   //下载地址
	decrypt           //解密方式
}

func (this *Response) Filename() string {
	return str.CropLast(this.filename, ".") + "ts"
}

type Url struct {
	url    string //资源地址
	isM3u8 bool   //是否是m3u8资源,否则是视频资源
}

type decrypt struct {
	Method string
	Key    []byte
	IV     []byte
}

func (this *decrypt) Decrypt(bs []byte) (_ []byte, err error) {
	defer g.Recover(&err)
	switch this.Method {
	case "AES-128":
		return aes.DecryptCBC(bs, this.Key, this.IV)
	}
	return bs, nil
}

/*




 */

func (this *Response) List() (list [][]*item, err error) {
	l := []*item(nil)
	for _, v := range this.TS_URL {
		if v.isM3u8 {
			resp, err := NewResponse(v.url)
			if err != nil {
				return nil, err
			}
			ls, err := resp.List()
			if err != nil {
				return nil, err
			}
			list = append(list, ls...)
			continue
		}
		l = append(l, &item{
			decrypt: this.decrypt,
			url:     v.url,
		})
	}
	if len(l) > 0 {
		list = append(list, l)
	}
	return
}

func NewTask(url string) ([]*task.Download, error) {
	resp, err := NewResponse(url)
	if err != nil {
		return nil, err
	}
	list, err := resp.List()
	if err != nil {
		return nil, err
	}
	tasks := []*task.Download(nil)
	for _, ls := range list {
		t := task.NewDownload()
		for _, v := range ls {
			t.Append(v)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

type item struct {
	decrypt
	url string
}

func (this *item) GetBytes(ctx context.Context, f func(p *http.Plan)) ([]byte, error) {
	bs, err := http.GetBytesWithPlan(this.url, f)
	if err != nil {
		return nil, err
	}
	return this.Decrypt(bs)
}
