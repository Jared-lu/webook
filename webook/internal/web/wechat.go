package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"webook/webook/internal/service"
	"webook/webook/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{svc: svc, userSvc: userSvc}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	hg := server.Group("/oauth2/wechat")
	hg.GET("/oauth2url", h.AuthURL)
	hg.Any("/callback", h.Callback)
}

// AuthURL 用于构造跳转到微信那边的URL
func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	_, err := h.svc.AuthURL(ctx.Request.Context(), state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造URL失败",
		})
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "构造URL成功",
	})
}

// Callback 处理从微信跳转回来的请求
func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	fmt.Println(state)
	info, err := h.svc.VerifyCode(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	user, err := h.userSvc.FindOrCreateByWechat(ctx.Request.Context(), info)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	// 保存登录态
	err = h.setJWTToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}
