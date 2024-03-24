package events

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"time"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/canalx"
	"webook/webook/pkg/logger"
	"webook/webook/pkg/saramax"
)

type MySQLBinlogConsumer struct {
	client sarama.Client
	l      logger.Logger
	cache  cache.FollowCache
}

func (r *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("follow_relations_cache",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			saramax.NewHandler[canalx.Message[dao.FollowRelation]](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage,
	val canalx.Message[dao.FollowRelation]) error {
	// 只监听 follow_relations 表
	if val.Table != "follow_relations" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Status {
		case dao.FollowRelationStatusActive:
			// 在缓存中创建关注关系
			err = r.cache.Follow(ctx, data.Follower, data.Followee)
		case dao.FollowRelationStatusInactive:
			// 在缓存中取消关注关系
			err = r.cache.CancelFollow(ctx, data.Follower, data.Followee)
		default:
			err = errors.New("unknown status")
		}
		if err != nil {
			return err
		}
	}
	return nil
}
