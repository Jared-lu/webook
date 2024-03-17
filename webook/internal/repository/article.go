package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"webook/webook/internal/domain"
	cache2 "webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/logger"
)

type CacheArticleRepository struct {
	dao dao.ArticleDAO

	// V1
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO
	// SyncV2 使用
	// 这意味着在这一层强耦合了DAO
	db *gorm.DB

	cache cache2.ArticleCache

	l logger.Logger
	// 组合UserRepository
	userRepo UserRepository
}

func (r *CacheArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	res, err := r.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.Article) domain.Article {
		return r.toDomain(src)
	}), nil
}

func (r *CacheArticleRepository) GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被你放过去了 OSS 上，你就要让前端去读 Content 字段
	art, err := r.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 你在这边要组装 user 了，适合单体应用
	usr, err := r.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.NickName,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (r *CacheArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成你的复杂的缓存方案
	// 你只缓存这一页
	if offset == 0 && limit <= 100 {
		data, err := r.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				r.preCache(ctx, data)
			}()
			//return data[:limit], err
			return data, err
		}
	}
	res, err := r.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return r.toDomain(src)
	})
	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err := r.cache.SetFirstPage(ctx, uid, data)
		r.l.Error("回写缓存失败", logger.Error(err))
		r.preCache(ctx, data)
	}()
	return data, nil
}

// preCache 预缓存
func (r *CacheArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := r.cache.Set(ctx, data[0])
		if err != nil {
			r.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}

func (r *CacheArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := r.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return r.toDomain(data), nil
}

func NewCacheArticleRepositoryV1(authorDAO dao.ArticleAuthorDAO, readerDAO dao.ArticleReaderDAO) ArticleRepository {
	return &CacheArticleRepository{authorDAO: authorDAO, readerDAO: readerDAO}
}

func NewCacheArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CacheArticleRepository{dao: dao}
}

func (r *CacheArticleRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error {
	return r.dao.SyncStatus(ctx, id, authorId, status.ToUint8())
}

// Sync 数据同步交给DAO层解决，在 repository 这一层认为只有一个DAO
func (r *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	// 同步制作库和线上库
	id, err := r.dao.Sync(ctx, r.toEntity(art))
	if err == nil {
		_ = r.cache.DelFirstPage(ctx, art.Author.Id)
		er := r.cache.SetPub(ctx, art)
		if er != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
	}
	return id, err
}

// SyncV2 尝试在 repository 层面上解决事务问题
// 这意味着repository要知道数据最终存在什么地方，是不是关系型数据库，是不是同库不同表等
// 这里制作库和线上库实际上是同库不同表
func (r *CacheArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 不能用tx1和tx2来操作两个不同的数据库事务，因为最后提交时要tx1.commit,tx2.commit，
	//可能tx1成功，但tx2不成功，也可能tx2成功，tx1不成功，即使使用原子操作也无法保证都成功或者都失败
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		// 开启事务失败，一般只有连接数不够才会触发
		return 0, tx.Error
	}
	defer tx.Rollback() // 如果commit成功，则这个语句会返回error，但不会有影响
	// 利用tx构建DAO
	// 这意味着实际上是在同一个数据库上执行操作，因此制作库和线上库实际上是同库不同表
	// 这样子还不如是把事务的操作下放到DAO层面执行，逼不得已都不要这样搞
	authorDAO := dao.NewAuthorDAO(tx)
	readerDAO := dao.NewReaderDAO(tx)
	var (
		id  = art.Id
		err error
	)
	artn := r.toEntity(art)
	if art.Id > 0 {
		err = authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		tx.Rollback() // 出错了，回滚事务
		return 0, err
	}
	// 不要忘了同步id
	artn.Id = id
	// 操作线上库
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle{Article: artn})
	tx.Commit() // 执行成功，提交事务
	return id, err
}

// SyncV1 无事务实现
// 制作库和线上库可以是不同库，也可以是同库不同表
func (r *CacheArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artn := r.toEntity(art)
	if art.Id > 0 {
		err = r.authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = r.authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	// 不要忘了同步id
	artn.Id = id
	// 操作线上库
	// 这里用的是同一张表，但不同库
	err = r.readerDAO.Upsert(ctx, r.toEntity(art))
	return id, err
}

func (r *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		r.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return r.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (r *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		r.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return r.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (r *CacheArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (r *CacheArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id: art.AuthorId,
		},
	}
}
