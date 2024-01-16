package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	cachemocks "webook/webook/internal/repository/cache/mocks"
	"webook/webook/internal/repository/dao"
	daomocks "webook/webook/internal/repository/dao/mocks"
)

func TestCacheUserRepository_FindById(t *testing.T) {
	// 这是纳秒级的时间，而数据库存的则是毫秒
	now := time.Now()
	// 去掉纳秒的部分
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name     string
		ctx      context.Context
		inputId  int64
		wantUser domain.User
		wantErr  error
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
	}{
		{
			name:    "未命中缓存，查询数据库成功",
			ctx:     context.Background(),
			inputId: 1,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				// 要手动转换为int64，否则默认为int
				d.EXPECT().FindById(context.Background(), int64(1)).Return(dao.User{
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "heelo@123",
					Phone: sql.NullString{
						String: "1312232332",
						Valid:  true,
					},
					NickName:    "测试",
					Birthday:    "1990-10-01",
					Description: "正在测试",
					Ctime:       now.UnixMilli(),
					Utime:       now.UnixMilli(),
				}, nil)
				c.EXPECT().Set(context.Background(), domain.User{
					Email:    "123@qq.com",
					Password: "heelo@123",
					Phone:    "1312232332",
					UserInfo: domain.UserInfo{
						NickName:    "测试",
						Birthday:    "1990-10-01",
						Description: "正在测试",
					},
					Ctime: now,
				}).Return(nil)
				return d, c
			},
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "heelo@123",
				Phone:    "1312232332",
				UserInfo: domain.UserInfo{
					NickName:    "测试",
					Birthday:    "1990-10-01",
					Description: "正在测试",
				},
				Ctime: now,
			},
			wantErr: nil,
		},
		{
			name:    "命中缓存",
			ctx:     context.Background(),
			inputId: 1,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{
					Email:    "123@qq.com",
					Password: "heelo@123",
					Phone:    "1312232332",
					UserInfo: domain.UserInfo{
						NickName:    "测试",
						Birthday:    "1990-10-01",
						Description: "正在测试",
					},
					Ctime: now}, nil)
				// 这一个测试用例不需要查询数据库
				return nil, c
			},
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "heelo@123",
				Phone:    "1312232332",
				UserInfo: domain.UserInfo{
					NickName:    "测试",
					Birthday:    "1990-10-01",
					Description: "正在测试",
				},
				Ctime: now,
			},
			wantErr: nil,
		},
		{
			name:    "查询数据库出错",
			ctx:     context.Background(),
			inputId: 1,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				// 要手动转换为int64，否则默认为int
				d.EXPECT().FindById(context.Background(), int64(1)).
					Return(dao.User{}, errors.New("数据库出错"))
				return d, c
			},
			wantUser: domain.User{},
			wantErr:  errors.New("数据库出错"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := NewUserRepository(tc.mock(ctrl))
			user, err := repo.FindById(tc.ctx, tc.inputId)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
			// 等待异步操作的执行
			//time.Sleep(time.Second)
		})
	}
}
