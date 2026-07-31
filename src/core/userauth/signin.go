package userauth

import (
	"apigw/src/core/proxy"
	"apigw/src/pkg/answer"
	"apigw/src/slog"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/sessions"
	"net/http"
	"strconv"
)

// UiasSignin 登录
func UiasSignin(host string, url string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		session := sessions.Default(c)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))
		if isLogin {
			c.JSON(200, answer.NewResMessage(answer.EcodeOkay, "User logged in.", nil))
			return
		}
		headers := proxy.SetReqHeaders(c)
		method := string(c.Method())
		body, _ := c.Body()

		// 发送http请求
		proxyPass := host + url

		response, err := proxy.HeProxy(ctx, method, "HTTP", "", proxyPass, body, headers)
		defer protocol.ReleaseResponse(response)
		if err != nil {
			klog.Error(err)
			c.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "The back-end service is abnormal.", nil))
			return
		}

		c.Status(response.StatusCode())
		response.Header.VisitAll(func(key, value []byte) {
			c.Response.Header.Set(string(key), string(value))
		})

		c.Response.Header.Del("X-Subject-Token")
		c.Response.Header.Del("X-Token-ExpireAt")

		if response.StatusCode() == 200 { // 登录成功后保存token
			// TODO 设置每个用户的单独会话时长
			klog.Info("log in success.")
			XSubjectToken := response.Header.Get("X-Subject-Token")
			session.Set("X-Subject-Token", XSubjectToken)
			session.Set("isLogin", true)
			if err = session.Save(); err != nil {
				klog.Errorf("session save err: %v", err)
				c.JSON(http.StatusInternalServerError,
					answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}
			klog.Info("log in success.")
		}
		c.Response.SetBody(response.Body())
	}
}
