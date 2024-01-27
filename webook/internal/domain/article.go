package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
}

// Author 文章作者
type Author struct {
	Id int64
	// 这里对应用户昵称
	Name string
}
