package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	web "webook/webook/internal/web/jwt"
	"webook/webook/pkg/logger"
)

// 业务
const biz = "login"

var _ handler = (*UserHandler)(nil)

// UserHandler 用户模块
type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	l           logger.Logger
	web.JWTHandler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl web.JWTHandler, l logger.Logger) *UserHandler {
	const (
		// 邮箱格式
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 密码格式
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	)
	// 使用预编译提高速度，好处是只需要编译一次
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		JWTHandler:  jwtHdl,
		l:           l,
	}
}

// RegisterRouter 注册用户模块路由
func (u *UserHandler) RegisterRouter(server *gin.Engine) {
	// 使用分组功能，自动在前面拼接/users
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	// 等价于  server.POST("/users/signup", u.SignUp)

	//ug.POST("/login", u.LoginJWT)
	//ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	//ug.POST("/login_sms", u.LoginSMS)
	//ug.POST("/refresh_token", u.RefreshToken)
	//ug.POST("/logout", u.LogoutJWT)

	// 不使用分组功能
	// server.POST("/users/login", u.Login)
	server.POST("/users/login", u.LoginJWTV1)
	//server.POST("/users/edit", u.Edit)
	server.POST("/users/edit", u.EditJWT)
	//server.GET("/users/profile", u.Profile)
	server.GET("/users/profile", u.ProfileJWT)
	server.POST("/users/login_sms/code/send", u.SendLoginSMSCode)
	server.POST("/users/login_sms", u.LoginSMS)
	server.POST("/users/refresh_token", u.RefreshToken)
	server.POST("/logout", u.LogoutJWT)
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	ctx.Next()
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
}

// RefreshToken 刷新短token
// 也可以同时刷新长短 token，用 redis 来记录长token是否有效，这时 refresh_token 是一次性的
// 可用参考登录校验部分，比较 User-Agent 来增强长token的安全性
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 只有这个接口，拿出来的才是 refresh_token，其它地方都是 access token
	//refreshToken := u.ExtractToken(ctx)
	// 校验token是否有效
	var claims web.RefreshClaims
	// 这里要保持传入结构体的指针
	err := u.CheckToken(ctx, &claims, web.RtKey)
	//token, err := jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
	//	// 我要解析的是长token
	//	return web.RtKey, nil
	//})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 长token需要检查是否已经退出登录了
	err = u.CheckSession(ctx, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 搞个新的 access_token，即只更新短token
	err = u.SetJWTToken(ctx, claims.Uid, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 可以在这里接着刷新长token

	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}

// SendLoginSMSCode 发送验证码
func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	// 校验手机号码是否合法
	// 这里可以使用一个正则表达式
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "手机号码不对",
		})
		return
	}

	err = u.codeSvc.Send(ctx.Request.Context(), biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
		return
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "发送太频繁",
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
}

// LoginSMS 验证码登录
func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 校验验证码
	ok, err := u.codeSvc.Verify(ctx.Request.Context(), biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "手机号码不对",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对",
		})
		return
	}

	// 登录或者注册，注册完还要保持登录
	// 新用户则直接创建，否则没有userId
	user, err := u.svc.FindOrCreateByPhone(ctx.Request.Context(), req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "验证码校验通过",
	})
}

// SignUp 注册
func (u *UserHandler) SignUp(ctx *gin.Context) {
	// 定义请求的参数结构
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}
	var req SignUpReq
	// 接收请求，拿到请求数据
	// Bind 方法会根据Content-Type 来解析数据，并写入到req中
	// 如果解析出错，就会写回一个400的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验邮箱格式
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误", // 不要直接返回具体的错误信息
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱格式不对",
		})
		return
	}
	// 校验密码格式
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "密码格式不对",
		})
		return
	}

	if req.ConfirmPassword != req.Password {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "前后两次输入的密码不匹配",
		})
		return
	}

	// 业务处理逻辑
	err = u.svc.SignUp(ctx.Request.Context(), domain.User{ // 不要直接传ctx
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱冲突",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 返回响应
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "注册成功",
	}) // 写回响应
}

// LoginJWTV1 使用带个人数据的JWT登录
func (u *UserHandler) LoginJWTV1(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidEmailOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 这里要用JWT保存登录态
	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}

// LoginJWT 使用JWT登录
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidEmailOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 这里要用JWT保存登录态
	// 生成JWT token
	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}

// Login 登录
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidEmailOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 这里要保存登录态
	// 拿到session
	sess := sessions.Default(ctx)
	// 设置session的值
	// 这个不是sess_id，是我们要存在Session里的数据
	// sess_id肯定是放在Cookie里
	// 那谁来生成这个sess_id？
	sess.Set("userId", user.Id) // 使用user id作为身份识别码
	// 控制Cookie
	sess.Options(sessions.Options{
		//Secure: true,
		MaxAge: 10 * 60, // 设置过期时间 30 * 60s（演示效果10*60s)
	})
	// 保存session
	sess.Save()
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}

// Logout 退出登录
func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 设置session的值（不是session_id)
	// 控制Cookie
	sess.Options(sessions.Options{
		// 让Cookie过期
		MaxAge: -1,
	})
	// 保存session
	sess.Save()
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "退出登录成功",
	})
}

func (u *UserHandler) EditJWT(ctx *gin.Context) {
	type EditReq struct {
		NickName    string `json:"name"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 这里缺了一些校验，如日期

	uid, _ := ctx.Get("userId")
	userId, ok := uid.(int64)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = u.svc.Edit(ctx.Request.Context(), domain.User{
		Id: userId,
		UserInfo: domain.UserInfo{
			NickName:    req.NickName,
			Birthday:    req.Birthday,
			Description: req.Description,
		},
	})
	if errors.Is(err, service.ErrInvalidEmailOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			// 可能是有人手动把用户的记录从数据库中删除了
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "编辑成功",
	})
}

// Edit 编辑个人信息
func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName    string `json:"name"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	sess := sessions.Default(ctx)
	id := sess.Get("userId")

	err = u.svc.Edit(ctx.Request.Context(), domain.User{
		Id: id.(int64),
		UserInfo: domain.UserInfo{
			NickName:    req.NickName,
			Birthday:    req.Birthday,
			Description: req.Description,
		},
	})
	if errors.Is(err, service.ErrInvalidEmailOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			// 可能是有人手动把用户的记录从数据库中删除了
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "编辑成功",
	})
}

// ProfileJWT 使用 JWT 机制
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	uid, _ := ctx.Get("userId")
	userId, ok := uid.(int64)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	user, err := u.svc.Profile(ctx.Request.Context(), domain.User{Id: userId})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	type ResProfile struct {
		NickName    string `json:"nickName"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	res := ResProfile{
		NickName:    user.NickName,
		Birthday:    user.Birthday,
		Description: user.Description,
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Data: res,
	})
}

// Profile 查看个人信息，使用Session机制
func (u *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	uid := sess.Get("userId")
	userId, _ := uid.(int64) // 类型断言，interface{}类型转换语法（明确知道是什么类型的情况下）
	user, err := u.svc.Profile(ctx.Request.Context(), domain.User{Id: userId})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	type ResProfile struct {
		NickName    string `json:"nickName"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	res := ResProfile{
		NickName:    user.NickName,
		Birthday:    user.Birthday,
		Description: user.Description,
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Data: res,
	})
}
