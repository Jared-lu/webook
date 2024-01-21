package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"webook/webook/internal/domain"
)

// 微信登录回调的URI
var redirectURI = url.PathEscape("https://your.com/oauth/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type LoginService struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewLoginService(appId string, appSecret string) Service {
	return &LoginService{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (w *LoginService) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = " https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, w.appId, redirectURI, state), nil
}

func (w *LoginService) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	targetURL := fmt.Sprintf(targetPattern, w.appId, w.appSecret, code)
	// 发起HTTP请求，验证code是否合法
	//resp, err := http.Get(targetURL)
	// 等价写法
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.Errcode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("微信返回错误响应，错误码: %d, 错误信息: %s", res.Errcode, res.Errmsg)
	}
	return domain.WechatInfo{
		OpenId:  res.Openid,
		UnionId: res.Unionid,
	}, nil
}

type Result struct {
	Errcode      int64  `json:"errcode"`
	Errmsg       string `json:"errmsg"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
	Unionid      string `json:"unionid"`
}
