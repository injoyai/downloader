package m3u8

import (
	"encoding/hex"
	"github.com/injoyai/base/bytes/crypt/aes"
	"github.com/injoyai/downloader/download"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

func RegexpAll(s string) []string {
	return regexp.MustCompile(`(http)[a-zA-Z0-9\\/=_\-.:%&]+\.m3u8([\?a-zA-Z0-9/=_\-.]{0,})`).FindAllString(s, -1)
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
			resp.TS_URL = append(resp.TS_URL, s)
			nextItem = false
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
		case strings.HasPrefix(s, "#EXTINF:"):
			//下一行是下载地址
			nextItem = true
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
	TS_URL   []string //下载地址
	decrypt           //解密方式
}

func (this *Response) Filename() string {
	return str.CropLast(this.filename, ".") + "ts"
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

func (this *Response) List() (list []download.Item) {
	for i, v := range this.TS_URL {
		list = append(list, &item{
			decrypt: this.decrypt,
			idx:     i,
			url:     v,
		})
	}
	return
}

func NewTask(url string) (download.Task, error) {
	resp, err := NewResponse(url)
	if err != nil {
		return nil, err
	}
	result := download.NewList(filepath.Join(resp.Filename()))
	for _, v := range resp.List() {
		result.Append(v)
	}
	return result, nil
}

type item struct {
	decrypt
	idx int
	url string
}

func (this *item) Idx() int {
	return this.idx
}

func (this *item) Run() ([]byte, error) {
	bs, err := http.GetBytes(this.url)
	if err != nil {
		return nil, err
	}
	return this.Decrypt(bs)
}
