package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
)

// 业务
const biz = "login"

// UserHandler 用户模块
type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
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
	server.POST("users/login_sms/code/send", u.SendLoginSMSCode)
	server.POST("users/login_sms", u.LoginSMS)
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

	err = u.codeSvc.Send(ctx, biz, req.Phone)
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
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
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
	user, err := u.svc.FindOrCreateByPhone(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = u.setJWTToken(ctx, user.Id)
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

// JWTUserClaims JWT用户数据
type JWTUserClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// 放入到token里的数据
	Uid       int64
	UserAgent string
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

	if err = u.setJWTToken(ctx, user.Id); err != nil {
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

func (u *UserHandler) setJWTToken(ctx *gin.Context, id int64) error {
	// 生成JWT token
	// JWT 带上个人数据作为一个身份识别
	claims := JWTUserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置jwt token的过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"))
	if err != nil {
		return err
	}
	// 将jwt token返回给前端，通过首部的方式
	ctx.Header("x-jwt-token", tokenStr)
	return nil
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
	token := jwt.New(jwt.SigningMethodHS512)
	tokenStr, err := token.SignedString([]byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 将jwt token返回给前端，通过首部的方式
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)

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
