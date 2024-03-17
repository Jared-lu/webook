package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
	"webook/webook/internal/domain"
	events "webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
)

type articleService struct {
	repo repository.ArticleRepository

	// V1 与上面互斥
	author repository.ArticleAuthorRepository
	reader repository.ArticleReaderRepository

	l logger.Logger

	producer events.Producer
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPublishedById(ctx *gin.Context, id int64, uid int64) (domain.Article, error) {
	// 另一个选项，在这里组装 Author，调用 UserService
	art, err := a.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			// 改批量的做法
			//a.ch <- readInfo{
			//	aid: id,
			//	uid: uid,
			//}
			er := a.producer.ProduceReadEvent(ctx, events.ReadEvent{
				Uid: uid,
				Aid: art.Id,
			})
			if er != nil {
				// 记录日志
			}
		}()

	}
	return art, err
}

func (a *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository,
	reader repository.ArticleReaderRepository, l logger.Logger) ArticleService {
	return &articleService{author: author, reader: reader, l: l}
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		// id > 0，说明不是新建，是编辑
		return art.Id, a.repo.Update(ctx, art)
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	//// 制作库
	//a.repo.Create(ctx, art)
	//// 同步到制作库
	//a.repo.SyncToLiveDB(ctx, art)
	return a.repo.Sync(ctx, art)
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
