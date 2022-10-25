package m3u8

import (
	"bytes"
	"encoding/hex"
	"github.com/injoyai/downloader/download"
	"github.com/injoyai/downloader/tool"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

func newKey(host, v string) (key *Key, err error) {
	v = tool.CropFirst(v, "#EXT-X-KEY:", false)
	key = new(Key)
	for _, k := range strings.Split(v, ",") {
		if l := strings.Split(k, "="); len(l) == 2 {
			switch l[0] {
			case "METHOD":
				key.method = l[1]
			case "URI":
				uri := l[1]
				uri = strings.ReplaceAll(uri, "\"", "")
				if !strings.HasPrefix(uri, "http") {
					uri = host + uri
				}
				key.key, err = tool.GetBytes(uri)
				if err != nil {
					return key, err
				}
			case "IV":
				if len(l[1]) > 2 && strings.ToLower(l[1][:2]) == "0x" {
					key.iv, err = hex.DecodeString(l[1][2:])
					if err != nil {
						return key, err
					}
				}
			}
		}
	}
	if len(key.iv) == 0 {
		key.iv = key.key
	}
	return
}

type Key struct {
	key    []byte
	method string
	iv     []byte
}

func (this *Key) decode(data []byte) (bytes []byte, err error) {
	defer tool.Recover(&err)
	if this == nil || len(this.key) == 0 {
		return data, nil
	}
	return tool.DecryptCBC(data, this.key, this.iv)
}

type Buff struct {
	start, end int
	*bytes.Buffer
}

type Info struct {
	Url   string
	bytes []byte
	err   error

	Key *Key
	idx int
}

func (this *Info) Idx() int {
	return this.idx
}

func (this *Info) Err() error {
	return this.err
}

func (this *Info) Bytes() []byte {
	return this.bytes
}

func (this *Info) Get() error {
	this.bytes, this.err = tool.GetBytes(this.Url)
	return this.err
}

func (this *Info) Run() ([]byte, error) {
	err := this.GetAndDecode()
	return this.Bytes(), err
}

func (this *Info) GetAndDecode() error {
	if err := this.Get(); err != nil {
		return err
	}
	this.bytes, this.err = this.Key.decode(this.bytes)
	return this.err
}

func (this *Info) Decode() (list []*Info, err error) {
	if err := this.Get(); err != nil {
		return nil, err
	}
	idx := 0
	for _, v := range strings.Split(string(this.bytes), "\n") {
		if strings.Contains(v, "#EXT-X-KEY:") {
			this.Key, err = newKey(this.prefix(), v)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(v, "http") {
			list = append(list, &Info{idx: idx, Url: v, Key: this.Key})
			idx++
		} else if strings.Contains(v, ".ts") || strings.Contains(v, ".png") {
			list = append(list, &Info{idx: idx, Url: this.prefix() + v, Key: this.Key})
			idx++
		} else if strings.Contains(v, ".m3u8") {
			if strings.Index(v, "/") > 1 {
				i := &Info{Url: this.prefix() + v}
				return i.Decode()
			} else {
				i := &Info{Url: this.host() + v}
				return i.Decode()
			}
		}
	}
	return
}

func (this *Info) Filename(name ...string) string {
	url := tool.CropLast(this.Url, "?", false)
	fileName := filepath.Base(url)
	fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), ".ts")
	if len(name) > 0 {
		fileName = name[0]
	}
	return fileName
}

func (this *Info) Filepath(name ...string) string {
	fileName := filepath.Base(this.Url)
	fileName = "./" + strings.ReplaceAll(fileName, filepath.Ext(fileName), ".ts")
	if len(name) > 0 {
		fileName = name[0]
	}
	return fileName
}

func (this *Info) prefix() string {
	return tool.CropLast(this.Url, "/")
}

func (this *Info) host() string {
	u, _ := url.Parse(this.Url)
	return u.Scheme + "://" + u.Host
}

func (this *Info) merge() (io.Reader, error) {

	list, err := this.Decode()
	if err != nil {
		return nil, err
	}
	result := make([]*Info, len(list))

	wg := sync.WaitGroup{}
	limit := make(chan byte, 20)
	for i, v := range list {
		wg.Add(1)
		limit <- 0
		go func(i int, v *Info) {
			defer func() {
				wg.Done()
				<-limit
			}()
			for x := 0; x < 3; x++ {
				if err := v.GetAndDecode(); err != nil {
					log.Println("错误:", err)
				} else {
					result[i] = v
					break
				}
			}
		}(i, v)
	}

	wg.Wait()
	buf := bytes.NewBuffer(nil)
	for _, v := range result {
		if v != nil {
			buf.Write(v.Bytes())
		}
	}

	return buf, nil
}

func (this *Info) Download(name ...string) error {
	fileName := this.Filepath(name...)
	buf, err := this.merge()
	if err != nil {
		return err
	}
	return tool.NewFile(fileName, buf)
}

func NewTask(url string) (download.Task, error) {
	x := Info{Url: url}
	list, err := x.Decode()
	if err != nil {
		return nil, err
	}
	result := download.NewList(x.Filename())
	for _, v := range list {
		result.Append(v)
	}
	return result, nil
}
