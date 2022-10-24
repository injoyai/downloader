package tool

import (
	"fmt"
	"strconv"
)

func NewBar() *Bar {
	return &Bar{
		prefix:  "进度:",
		suffix:  "",
		length:  40,
		nowSize: 0,
		maxSize: 1000,
		style:   ">",
		//color:   0,
		c:     make(chan string, 1),
		done:  make(chan uintptr, 1),
		print: func(s string) { fmt.Print(s) },
	}
}

type Bar struct {
	prefix  string  //前缀
	suffix  string  //后缀
	length  int     //总长度
	nowSize float64 //当前完成数量
	maxSize float64 //最大数量
	style   string  //进度条风格
	//color   int  //整体颜色
	c     chan string  //实时数据通道
	done  chan uintptr //结束信号
	print func(string) //打印
}

// SetPrint 设置打印函数
func (this *Bar) SetPrint(fn func(string)) *Bar {
	this.print = fn
	return this
}

// SetPrefix 设置前缀
func (this *Bar) SetPrefix(prefix string) *Bar {
	this.prefix = prefix
	return this
}

// SetLength 设置进度条长度
func (this *Bar) SetLength(length int) *Bar {
	this.length = length
	return this
}

// SetSize 设置进度任务数量
func (this *Bar) SetSize(size float64) *Bar {
	this.maxSize = size
	return this
}

// SetStyle 设置进度条风格
func (this *Bar) SetStyle(style string) *Bar {
	this.style = style
	return this
}

// SetColor 设置进度条颜色
func (this *Bar) SetColor(color int) *Bar {
	//this.color = color
	return this
}

func (this *Bar) Done() {
	this.Add(this.maxSize - this.nowSize)
}

func (this *Bar) Add(n float64) {
	this.nowSize += n
	if this.nowSize >= this.maxSize {
		this.nowSize = this.maxSize
		defer func() {
			select {
			case <-this.done:
			default:
				close(this.done)
			}
		}()
	}
	nowLength := int((this.nowSize / this.maxSize) * float64(this.length) / float64(len(this.style)))
	s := ""
	for i := 0; i < nowLength; i++ {
		s += this.style
	}
	this.c <- s
}

func (this *Bar) Print(v ...interface{}) {
	fmt.Printf("\r%s", fmt.Sprint(v...))
}

func (this *Bar) Wait() <-chan uintptr {
	this.Add(0)
	for {
		select {
		case <-this.done:
			fmt.Println("")
			return this.done
		case s := <-this.c:
			width := strconv.Itoa(this.length)
			//if this.color > 0 {
			//s = color.Set(this.color, s)
			width = strconv.Itoa(this.length + 10)
			//}
			s = fmt.Sprintf("\r%s[%-"+width+"s] %0.1f%% %0.0f/%0.0f %s", this.prefix, s, this.nowSize*100/this.maxSize, this.nowSize, this.maxSize, this.suffix)
			if this.print != nil {
				this.print(s)
			} else {
				fmt.Print(s)
			}
		}
	}
}
