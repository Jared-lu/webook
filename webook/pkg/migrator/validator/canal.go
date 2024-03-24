package validator

import (
	"context"
	"gorm.io/gorm"
	"webook/webook/pkg/logger"
	"webook/webook/pkg/migrator"
	"webook/webook/pkg/migrator/events"
)

type CanalIncrValidator[T migrator.Entity] struct {
	baseValidator
}

func NewCanalIncrValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.Logger,
	producer events.Producer,
) *CanalIncrValidator[T] {
	return &CanalIncrValidator[T]{
		baseValidator: baseValidator{
			base:      base,
			target:    target,
			direction: direction,
			l:         l,
			producer:  producer,
		},
	}
}

// Validate 一次校验一条
func (v *CanalIncrValidator[T]) Validate(ctx context.Context, id int64) error {
	var base T
	err := v.base.WithContext(ctx).Where("id = ?", id).First(&base).Error
	switch err {
	case nil:
		// 找到了
		var target T
		err1 := v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err1 {
		case nil:
			// target 里面也找到了
			if !base.CompareTo(target) {
				v.notify(id, events.InconsistentEventTypeNEQ)
			}
		case gorm.ErrRecordNotFound:
			v.notify(id, events.InconsistentEventTypeTargetMissing)
		default:
			return err
		}
	case gorm.ErrRecordNotFound:
		// 找到了
		var target T
		err1 := v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err1 {
		case nil:
			// target 里面也找到了
			v.notify(id, events.InconsistentEventTypeBaseMissing)
		case gorm.ErrRecordNotFound:
			// 两边都没了，啥也不需要干
		default:
			return err
		}
	// 收到消息的时候，或者说收到 binlog 的时候，这条数据已经没了
	default:
		return err
	}
	return nil
}