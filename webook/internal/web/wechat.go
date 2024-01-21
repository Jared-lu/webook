package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
	"webook/webook/internal/service"
	"webook/webook/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
	stateKey []byte
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:        svc,
		userSvc:    userSvc,
		stateKey:   []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"),
		jwtHandler: newJwtHandler(),
	}
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
		return
	}
	// 保存state，从微信回来后要进行校验
	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "构造URL成功",
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			// 预期一个用户完成登录的时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 3)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600,
		"/oauth2/wechat/callback", "", false, true)
	return nil
}

// Callback 处理从微信跳转回来的请求
func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	// 校验state
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "登录失败",
		})
		return
	}
	info, err := h.svc.VerifyCode(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	user, err := h.userSvc.FindOrCreateByWechat(ctx.Request.Context(), info)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 保存登录态
	err = h.setJWTToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.setRefreshToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	// 校验state
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("拿不到state的cookie, %w", err)
	}

	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token已过期, %w", err)
	}
	if sc.State != state {
		return fmt.Errorf("state不相等, %w", err)
	}
	return nil
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
