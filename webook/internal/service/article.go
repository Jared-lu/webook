package service

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
)

type articleService struct {
	repo repository.ArticleRepository

	// V1 与上面互斥
	author repository.ArticleAuthorRepository
	reader repository.ArticleReaderRepository

	l logger.Logger
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository,
	reader repository.ArticleReaderRepository, l logger.Logger) ArticleService {
	return &articleService{author: author, reader: reader, l: l}
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		// id > 0，说明不是新建，是编辑
		return art.Id, a.repo.Update(ctx, art)
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	//// 制作库
	//a.repo.Create(ctx, art)
	//// 同步到制作库
	//a.repo.SyncToLiveDB(ctx, art)
	return a.repo.SyncV1(ctx, art)
}
func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 确保制作库和线上库的值相等
	art.Id = id
	// 接下来无法区分编辑再发表和重新发表的情况，因此直接存到Repository上，让Repository去解决update or insert
	for i := 0; i < 3; i++ {
		// 重试
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
	}
	if err != nil {
		a.l.Error("重试全部失败",
			logger.Int64("art_id", art.Id), logger.Error(err))
	}
	return id, err
}
