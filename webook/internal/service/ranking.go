package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

// RankingService 热榜计算
type RankingService interface {
	TopN(ctx context.Context) error
	//TopN(ctx context.Context, n int64) error
	//TopN(ctx context.Context, n int64) ([]domain.Article, error)
}

// BatchRankingService 批量异步定时计算热榜
type BatchRankingService struct {
	// 为了拿到一批文章
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) *BatchRankingService {
	return &BatchRankingService{artSvc: artSvc, intrSvc: intrSvc}
}

// TopN 准备分批计算，每批的数量为N
func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		// 在这里先存起来
	}
	return svc.repo.ReplaceTopN(ctx, arts)
}

// 为了topN 剥离测试
func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 先拿一批
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		// 拿到这批文章的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 合并计算score
		// 排序
		for _, art := range arts {
			intr := intrs[art.Id]
			//if !ok {
			//	// 你都没有，肯定不可能是热榜
			//	continue
			//}
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			// 我要考虑，我这个 score 在不在前一百名
			// 拿到热度最低的
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			// 这种写法，要求 topN 已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					_ = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}
		// 一批已经处理完了，问题来了，我要不要进入下一批？我怎么知道还有没有？
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 我这一批都没取够，我当然可以肯定没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 这边要更新 offset
		offset = offset + len(arts)
	}
	// 最后得出结果
	res := make([]domain.Article, svc.n)
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明取完了，不够 n
			break
		}
		res[i] = val.art
	}
	return res, nil
}
