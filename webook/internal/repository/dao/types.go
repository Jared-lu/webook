package dao

import (
	"context"
	"time"
)

//go:generate mockgen -source=./types.go -package=daomocks -destination=./mocks/dao.mock.go

type UserDAO interface {
	FindById(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Insert(ctx context.Context, u User) error
	InsertV1(ctx context.Context, u User) (User, error)
	Update(ctx context.Context, u User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechatOpenId(ctx context.Context, openId string) (User, error)
	Search(ctx context.Context, keywords []string) ([]User, error)
	InputUser(ctx context.Context, u User) error
}

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
	GetByAuthor(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error)
	Search(ctx context.Context, likeIds []int64, collectIds []int64, tagIds []int64, keywords []string) ([]Article, error)
	InputArticle(ctx context.Context, articl Article) error
}

type ArticleAuthorDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}

type ArticleReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

type CollectDAO interface {
	InputCollect(ctx context.Context, collect Collect) error
	Search(ctx context.Context, uid int64, biz string) ([]int64, error)
}

type LikeDAO interface {
	InputLike(ctx context.Context, like Like) error
	Search(ctx context.Context, uid int64, biz string) ([]int64, error)
}

type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelationDao interface {
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}
