package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
)

// UserHandler 用户模块
type UserHandler struct {
	svc         service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService) *UserHandler {
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
	server.POST("/users/login", u.Login)
	server.POST("/users/edit", u.Edit)
	server.GET("/users/profile", u.Profile)
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好")
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
	sess.Set("userId", user.Id)
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

// Edit 编辑个人信息
func (u *UserHandler) Edit(ctx *gin.Context) {

}

// Profile 查看个人信息
func (u *UserHandler) Profile(ctx *gin.Context) {

}
