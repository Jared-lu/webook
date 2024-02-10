package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

// Author 文章作者
type Author struct {
	Id int64
	// 这里对应用户昵称
	Name string
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(8)
}

func (s ArticleStatus) Valid() bool {
	return s.ToUint8() > 0
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusUnpublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	case ArticleStatusPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// ArticleStatusV1 对于十分复杂的状态，如有很多方法，	可以是用这种形态
type ArticleStatusV1 struct {
	Val    uint8
	String string
}
