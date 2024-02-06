package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/webook/internal/integration/startup"
	"webook/webook/internal/repository/dao"
)

type ArticleHandlerTestSuite struct {
	// 使用测试套件，减少冗余代码
	suite.Suite
	server *gin.Engine
	// 这个db是用来验证数据库的
	db *gorm.DB
}

// TestArticle 使用测试套件必不可少的一步
// 运行这个方法会运行所注册的全部测试套件的测试方法
func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleHandlerTestSuite{})
}

func (s *ArticleHandlerTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	//s.server = startup.InitApp()
	s.server = gin.Default()
	// 我这里没有使用其它的中间件
	s.server.Use(func(ctx *gin.Context) {
		// 测试可以直接设置一个用户登录态
		ctx.Set("userId", int64(123))
	})
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRouter(s.server)
}

func (s *ArticleHandlerTestSuite) TearDownTest() {
	// 清空所有数据库，并将自增主键恢复为1
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	assert.NoError(s.T(), err)
	s.db.Exec("TRUNCATE TABLE `published_articles`")
}

func (s *ArticleHandlerTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		// 这个是我输入的文章
		art      Article
		wantCode int
		// 预期返回的是文章的Id，因此直接用int64
		wantBody Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {
				// 什么都不用做
			},
			after: func(t *testing.T) {
				var art dao.Article
				// 拿到我要新建的那篇帖子，这里我事先知道了我插入的文章ID就是1
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				// 与时间戳有关的比较技巧
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				// 断言取出来的数据和我预期的相同
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的帖子",
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的帖子",
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{
				// 返回文章的ID
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    1234,
					Utime:    1234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art dao.Article
				// 拿到我要新建的那篇帖子，这里我事先知道了我插入的文章ID就是1
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 1234)
				art.Utime = 0
				// 断言取出来的数据和我预期的相同
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    1234,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{
				// 返回文章的ID
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Ctime:    1234,
					Utime:    1234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art dao.Article
				// 拿到我要新建的那篇帖子，这里我事先知道了我插入的文章ID就是1
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				// 断言取出来的数据和我预期的相同
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的标题",
					AuthorId: 789,
					Ctime:    1234,
					Utime:    1234,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{
				// 返回文章的ID
				Data: 5,
				Msg:  "保存失败",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			// 构造http请求
			body, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			// 构造响应
			resp := httptest.NewRecorder()
			// 启动服务器
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				// 400响应码没有必要笔记下面了
				return
			}
			var result Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)
			tc.after(t)
		})
	}
}

func (s *ArticleHandlerTestSuite) TestArticleHandler_Publish() {
	panic("implement me")
}

// Article 预期中的article输入，测试用的
type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Result 响应体中Data是any，存在序列化问题，不好比较
type Result[T any] struct {
	// 业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
