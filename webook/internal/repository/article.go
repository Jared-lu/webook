package repository

import (
	"context"
	"gorm.io/gorm"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao"
)

type CacheArticleRepository struct {
	dao dao.ArticleDAO

	// V1
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO
	// SyncV2 使用
	// 这意味着在这一层强耦合了DAO
	db *gorm.DB
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
	return r.dao.Sync(ctx, r.toEntity(art))
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
	return r.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (r *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
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
