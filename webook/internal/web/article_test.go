package web

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"webook/webook/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		reqBody  string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		wantCode int
		wantBody Result
	}{
		{
			name: "新建帖子，发表成功",
			reqBody: `
	{
		"title": "我的标题",
		"content": "我的内容"
	}
`,
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 789,
					},
				}).Return(int64(1), nil)
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				// json反序列化时，数值默认是float64
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "发表失败",
			reqBody: `
	{
		"title": "我的标题",
		"content": "我的内容"
	}
`,
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 789,
					},
				}).Return(int64(0), errors.New("发表失败"))
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
			// 模拟登录态
			server.Use(func(ctx *gin.Context) {
				ctx.Set("userId", int64(789))
			})
			u := NewArticleHandler(tc.mock(ctrl), logger.NewNoOpLogger())
			u.RegisterRouter(server)
			// 构造http请求
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
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
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)
		})
	}
}
