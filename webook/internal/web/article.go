package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	web "webook/webook/internal/web/jwt"
	"webook/webook/pkg/ginx"
	"webook/webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	l       logger.Logger
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, l: l}
}

func (h *ArticleHandler) RegisterRouter(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// 创作者的查询接口
	// 这个是获取数据的接口，理论上来说（遵循 RESTful 规范），应该是用 GET 方法
	// GET localhost/articles => List 接口
	g.POST("/list",
		ginx.WrapBodyAndToken[ListReq, web.JWTUserClaims](h.List))
	g.GET("/detail/:id", ginx.WrapToken[web.JWTUserClaims](h.Detail))

	// 读者的阅读文章
	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail, func(ctx *gin.Context) {
		// 增加阅读计数。
		//go func() {
		//	// 开一个 goroutine，异步去执行
		//	er := a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
		//	if er != nil {
		//		a.l.Error("增加阅读计数失败",
		//			logger.Int64("aid", art.Id),
		//			logger.Error(err))
		//	}
		//}()
	})

	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//pub.POST("/like/:id", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq,
		web.JWTUserClaims](h.Like))
}

func (a *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc web.JWTUserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = a.intrSvc.Like(ctx, a.biz, req.Id, uc.Uid)
	} else {
		err = a.intrSvc.CancelLike(ctx, a.biz, req.Id, uc.Uid)
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	uc := ctx.MustGet("users").(web.JWTUserClaims)
	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = a.svc.GetPublishedById(ctx, id, uc.Uid)
		return err
	})

	//var getResp *intrv1.GetResponse
	//eg.Go(func() error {
	//	// 这个地方可以容忍错误
	//	getResp, err = a.intrSvc.Get(ctx, &intrv1.GetRequest{
	//		Biz: a.biz, BizId: id, Uid: uc.Id,
	//	})
	//	// 这种是容错的写法
	//	//if err != nil {
	//	//	// 记录日志
	//	//}
	//	//return nil
	//	return err
	//})

	// 在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 增加阅读计数。
	go func() {
		// 你都异步了，怎么还说有巨大的压力呢？
		// 开一个 goroutine，异步去执行
		er := a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
		if er != nil {
			a.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		}
	}()

	// ctx.Set("art", art)

	//intr := getResp.Intr

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author: art.Author.Name,
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
			//Liked:      intr.Liked,
			//Collected:  intr.Collected,
			//LikeCnt:    intr.LikeCnt,
			//ReadCnt:    intr.ReadCnt,
			//CollectCnt: intr.CollectCnt,
		},
	})
}

// Detail 文章详情接口
func (a *ArticleHandler) Detail(ctx *gin.Context, usr web.JWTUserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != usr.Uid {
		//ctx.JSON(http.StatusOK)
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		//a.l.Error("非法访问文章，创作者 ID 不匹配",
		//	logger.Int64("uid", usr.Id))
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
}

// List 创作者的文章列表接口
func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc web.JWTUserClaims) (ginx.Result, error) {
	res, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uid, ok := ctx.Get("userId")
	userId, ok := uid.(int64)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}
	err := h.svc.Withdraw(ctx.Request.Context(), domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: userId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}
func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uid, ok := ctx.Get("userId")
	userId, ok := uid.(int64)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}

	id, err := h.svc.Publish(ctx.Request.Context(), domain.Article{
		// 新建并发表时是没有ID的，为0
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: userId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
		Msg:  "OK",
	})
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 验证输入

	uid, ok := ctx.Get("userId")
	userId, ok := uid.(int64)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: userId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "保存失败",
		})
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
		Msg:  "OK",
	})
}
