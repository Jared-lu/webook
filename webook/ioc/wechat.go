package ioc

import "webook/webook/internal/service/oauth2/wechat"

func InitOAuth2WechatService() wechat.Service {
	appId := ""
	appSecret := ""
	return wechat.NewLoginService(appId, appSecret)
}
