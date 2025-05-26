package userauth

import (
	"apigw/src/pkg/answer"
	"apigw/src/pkg/core"
	"apigw/src/slog"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
)

// UiasSignin 登录
func UiasSignin(host string, url string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)
		session := sessions.Default(c)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))
		if isLogin {
			c.JSON(200, answer.NewResMessage(answer.EcodeOkay, "User logged in.", nil))
			return
		}

		headers := core.SetReqHeaders(ctx, c)
		method := string(c.Method())
		body, _ := c.Body()

		// 发送http请求
		proxyPass := host + url

		response, err := core.SendHttpRequest(ctx, method, proxyPass, body, headers, 10*time.Second)
		if err != nil {
			klog.Error(err)
			c.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "The back-end service is abnormal.", nil))
			return
		}

		// 设置响应头
		for key, value := range response.Header {
			klog.Debugf("Header[%q] = %q", key, value)
			c.Response.Header.Set(key, strings.Join(value, ""))
		}

		// 删除不必要的头部
		c.Response.Header.Del("X-Subject-Token")
		c.Response.Header.Del("X-Token-ExpireAt")

		if response.StatusCode == 200 {
			// 登录成功后保存token
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

		// 无论成功或失败都透传上游的响应
		c.Data(response.StatusCode, response.Header.Get("Content-Type"), response.Body)
	}
}
