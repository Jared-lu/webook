package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	svcmocks "webook/webook/internal/service/mocks"
)

func TestTypeAssert(t *testing.T) {
	var id any
	if id == nil {
		fmt.Println("nil")
	}
	// 当变量为nil时，类型断言失败，但不会panic
	_, ok := id.(int64)
	if ok {
		fmt.Println("ok")
	} else {
		fmt.Println("!ok")
	}
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		/* 每个测试用例特有的则抽取出来 */
		name string
		// 用来模拟需要用到的第三方组件
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		// 用string的话不好判断，直接判等结构体更方便
		wantBody Result
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 预期第三方组件会有做些什么
				userSvc.EXPECT().SignUp(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(nil)
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123",
	"confirmPassword":"hello@123"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 0,
				Msg:  "注册成功",
			},
		},
		{
			name: "Bind方法失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123",
	
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email":"1234",
	"password":"hello@123",
	"confirmPassword":"hello@123"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "邮箱格式不对",
			},
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello123",
	"confirmPassword":"hello123"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "密码格式不对",
			},
		},
		{
			name: "前后两次输入的密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123",
	"confirmPassword":"hello@12345"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "前后两次输入的密码不匹配",
			},
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 预期第三方组件会有做些什么
				userSvc.EXPECT().SignUp(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(service.ErrUserDuplicateEmail)
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123",
	"confirmPassword":"hello@123"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "邮箱冲突",
			},
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 预期第三方组件会有做些什么
				userSvc.EXPECT().SignUp(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(errors.New("随便一个错误"))
				return userSvc
			},
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123",
	"confirmPassword":"hello@123"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 公共的部分放这里面
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 和正常使用一样，都需要先初始化服务器和UserHandler等操作
			server := gin.Default()
			// Signup接口不需要用到验证码服务
			u := NewUserHandler(tc.mock(ctrl), nil)
			u.RegisterRouter(server)
			// 构造http请求
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			// 构造响应
			resp := httptest.NewRecorder()
			// 启动服务器
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				// 400响应码没有必要笔记下面了
				return
			}
			var result Result
			err = json.Unmarshal(resp.Body.Bytes(), &result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)

		})
	}
}

func TestUserHandler_LoginJWTV1(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody Result
	}{
		{
			name: "登录成功",
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123"
}
`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Login(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(domain.User{}, nil)
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 0,
				Msg:  "登录成功",
			},
		},
		{
			name: "请求有误",
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@12
`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			wantCode: http.StatusBadRequest,
			wantBody: Result{},
		},
		{
			name: "邮箱或密码不对",
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123"
}
`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Login(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(domain.User{}, service.ErrInvalidEmailOrPassword)
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "邮箱或密码不对",
			},
		},
		{
			name: "系统错误",
			reqBody: `
{
	"email":"1234@qq.com",
	"password":"hello@123"
}
`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Login(context.Background(), domain.User{
					Email:    "1234@qq.com",
					Password: "hello@123",
				}).Return(domain.User{}, errors.New("系统错误"))
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			u := NewUserHandler(tc.mock(ctrl), nil)
			u.RegisterRouter(server)
			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				// 400响应码没有必要笔记下面了
				return
			}
			var res Result
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name     string
		reqBody  string
		mock     func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		wantCode int
		wantBody Result
	}{
		{
			name: "登录成功",
			reqBody: `
{
    "phone":"13761234565",
    "code":"355673"
}
`,
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(context.Background(), biz, "13011223344", "012345").
					Return(true, nil)
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().FindOrCreateByPhone(context.Background(), "13011223344").
					Return(domain.User{}, nil)
				return userSvc, codeSvc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 0,
				Msg:  "验证码校验通过"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//ctrl := gomock.NewController(t)
			//ctrl.Finish()
			//server := gin.Default()
			//u := NewUserHandler(tc.mock(ctrl))
			//u.RegisterRouter(server)
			//req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			//require.NoError(t, err)
			//req.Header.Set("Content-Type", "application/json")
			//resp := httptest.NewRecorder()
			//server.ServeHTTP(resp, req)
			//assert.Equal(t, tc.wantCode, resp.Code)
			//if resp.Code != 200 {
			//	return
			//}
			//var res Result
			//err = json.Unmarshal(resp.Body.Bytes(), &res)
			//require.NoError(t, err)
			//assert.Equal(t, tc.wantBody, res)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			u := NewUserHandler(tc.mock(ctrl))
			u.RegisterRouter(server)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				// 400响应码没有必要笔记下面了
				return
			}
			var res Result
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
