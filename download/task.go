package download

type Task interface {
	Filename() string
	Len() int
	Next() Item
	List() []Item
}

type Item interface {
	Idx() int
	Run() ([]byte, error)
}

type List struct {
	filename string
	list     []Item
}

func (this *List) Append(i Item) {
	this.list = append(this.list, i)
}

func (this *List) Filename() string {
	return this.filename
}

func (this *List) Next() Item {
	if len(this.list) == 0 {
		return nil
	}
	item := this.list[0]
	this.list = this.list[1:]
	return item
}

func (this *List) List() []Item {
	return this.list
}

func (this *List) Len() int {
	return len(this.list)
}

func NewList(filename string) *List {
	return &List{
		filename: filename,
	}
}
