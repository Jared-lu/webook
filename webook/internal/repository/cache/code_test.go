package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	redismock2 "webook/webook/mock/redis"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
		mock    func(ctrl *gomock.Controller) redis.Cmdable
	}{
		{
			name:    "设置验证码成功",
			ctx:     context.Background(),
			biz:     "login",
			phone:   "13722223333",
			code:    "000222",
			wantErr: nil,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				r := redismock2.NewMockCmdable(ctrl)
				// 控制返回值，模拟Int()方法
				res := redis.NewCmdResult(int64(0), nil)
				r.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:13722223333"}, "000222").
					Return(res)
				return r
			},
		},
		{
			name:    "发送太频繁",
			ctx:     context.Background(),
			biz:     "login",
			phone:   "13722223333",
			code:    "000222",
			wantErr: ErrCodeSendTooMany,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				r := redismock2.NewMockCmdable(ctrl)
				// 控制返回值，模拟Int()方法
				res := redis.NewCmdResult(int64(-1), nil)
				r.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:13722223333"}, "000222").
					Return(res)
				return r
			},
		},
		{
			name:    "设置验证码成功",
			ctx:     context.Background(),
			biz:     "login",
			phone:   "13722223333",
			code:    "000222",
			wantErr: errors.New("系统错误"),
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				r := redismock2.NewMockCmdable(ctrl)
				// 控制返回值，模拟Int()方法
				res := redis.NewCmdResult(int64(-2), nil)
				r.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:13722223333"}, "000222").
					Return(res)
				return r
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctrl.Finish()
			r := NewRedisCodeCache(tc.mock(ctrl))
			err := r.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
