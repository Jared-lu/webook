package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	repomocks "webook/webook/internal/repository/mocks"
)

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name      string
		inputUser domain.User
		wantUser  domain.User
		wantErr   error
		mock      func(ctrl *gomock.Controller) repository.UserRepository
	}{
		{
			name: "登录成功",
			inputUser: domain.User{
				Email: "123@qq.com",
				// 这里要输入原文
				Password: "hello@123",
			},
			wantUser: domain.User{
				Email: "123@qq.com",
				// 这里是加密后的密码
				Password: "$2a$10$NK/Esc3Etd2Yamqt/4OjK.88gFapQyBo2iWrn7z2uTHGUlsiffyiq",
			},
			wantErr: nil,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{
					Email: "123@qq.com",
					// 这里是加密后的密码
					Password: "$2a$10$NK/Esc3Etd2Yamqt/4OjK.88gFapQyBo2iWrn7z2uTHGUlsiffyiq",
				}, nil)
				return repo
			},
		},
		{
			name: "用户不存在",
			inputUser: domain.User{
				Email: "123@qq.com",
				// 这里要输入原文
				Password: "hello@123",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidEmailOrPassword,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
		},
		{
			name: "系统异常",
			inputUser: domain.User{
				Email: "123@qq.com",
				// 这里要输入原文
				Password: "hello@123",
			},
			wantUser: domain.User{},
			wantErr:  errors.New("系统或数据库异常错误"),
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{}, errors.New("系统或数据库异常错误"))
				return repo
			},
		},
		{
			name: "密码不匹配",
			inputUser: domain.User{
				Email: "123@qq.com",
				// 这里要输入错误的密码
				Password: "123hello@123",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidEmailOrPassword,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{
					Email: "123@qq.com",
					// 返回的依然是正确密码的密文
					Password: "$2a$10$NK/Esc3Etd2Yamqt/4OjK.88gFapQyBo2iWrn7z2uTHGUlsiffyiq",
				}, nil)
				return repo
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Login(context.Background(), tc.inputUser)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncrypt(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("hello@123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(hash))
	}
}
