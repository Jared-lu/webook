package repository

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao"
)

type collectRepository struct {
	dao dao.CollectDAO
}

func (c *collectRepository) InputCollect(ctx context.Context, msg domain.Collect) error {
	return c.dao.InputCollect(ctx, dao.Collect{
		Uid:   msg.Uid,
		Biz:   msg.Biz,
		BizId: msg.BizId,
	})
}

func NewCollectRepository(d dao.CollectDAO) CollectRepository {
	return &collectRepository{
		dao: d,
	}
}
