package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/webook/internal/integration/startup"
	"webook/webook/internal/web"
)

func TestUserHandler_SendLoginSMSCode(t *testing.T) {
	// 初始化需要用到的组件和服务器
	server := startup.InitApp()
	rdb := startup.InitRedis()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送验证码成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// 在指定时间内删除数据，否则认为出错
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理请求放入的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:13712345679").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)

			},
			reqBody: `
{
	"phone": "13712345679"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "请求有误",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			reqBody: `
{
	"phone": 
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "手机号不对",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			reqBody: `
{
	"phone": ""
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "手机号码不对",
			},
		},
		{
			name: "发送验证码太频繁",
			before: func(t *testing.T) {
				_, err := rdb.Set(context.Background(), "phone_code:login:13712345679", "123456",
					// 必须要在一分钟内
					time.Minute*9+time.Second*30).Result()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 在指定时间内删除数据，否则认为出错
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理请求放入的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:13712345679").Result()
				cancel()
				assert.NoError(t, err)
				// 比较一下验证码对不对
				assert.Equal(t, "123456", val)

			},
			reqBody: `
{
	"phone": "13712345679"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "发送太频繁",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				_, err := rdb.Set(context.Background(), "phone_code:login:13712345679", "123456",
					// 过期时间为0就是没有过期时间
					0).Result()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 在指定时间内删除数据，否则认为出错
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理请求放入的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:13712345679").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)

			},
			reqBody: `
{
	"phone": "13712345679"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			// 构造http请求
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
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
			var result web.Result
			err = json.Unmarshal(resp.Body.Bytes(), &result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)
			tc.after(t)
		})
	}
}
