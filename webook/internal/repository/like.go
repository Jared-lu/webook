package repository

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao"
)

type likeRepository struct {
	dao dao.LikeDAO
}

func (l *likeRepository) InputLike(ctx context.Context, msg domain.Like) error {
	return l.dao.InputLike(ctx, dao.Like{
		Uid:    msg.Uid,
		Biz:    msg.Biz,
		BizId:  msg.BizId,
		Status: msg.Status,
	})
}

func NewLikeRepository(d dao.LikeDAO) LikeRepository {

	return &likeRepository{dao: d}
}
