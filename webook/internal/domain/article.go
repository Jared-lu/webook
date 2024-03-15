package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

// MarkedAsPrivate 将文章状态设置为仅自己可见
func (a *Article) MarkedAsPrivate() {
	a.Status = ArticleStatusPrivate
}

// Abstract 提取文章摘要
func (a Article) Abstract() string {
	// 摘要我们取前几句。
	// 要考虑一个中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	// 英文怎么截取一个完整的单词，我的看法是……不需要纠结，就截断拉到
	// 词组、介词，往后找标点符号
	return string(cs[:100])
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
