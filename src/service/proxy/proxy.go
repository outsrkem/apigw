package proxy

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

// SetUrl 去除url前缀，并设置请求到后端的完整url
func SetUrl(c *app.RequestContext, target, rUrl string) string {
	return target + strings.TrimPrefix(string(c.URI().RequestURI()), rUrl)
}

// StartProxy 反向代理
func StartProxy(ctx context.Context, c *app.RequestContext, session sessions.Session, target, path string) {
	klog := slog.FromContext(c)
	headers := core.SetReqHeaders(ctx, c) // 头部处理, 获取原有请求头并透传
	method := string(c.Method())
	body, err := c.Body()
	if err != nil {
		klog.Errorf("body err: %v", err)
	}
	klog.Debugf("req body: %v", string(body))

	// TODO 移除频繁session的保存
	if XSubjectToken, ok := session.Get("X-Subject-Token").(string); ok {
		headers["X-Auth-Token"] = XSubjectToken       // 转换成功，可以使用
		session.Set("X-Subject-Token", XSubjectToken) // 保存一次session，防止过期
		if err := session.Save(); err != nil {
			klog.Errorf("session save error: %v", err)
		}
	}

	// TODO token 续签
	url := SetUrl(c, target, path)
	klog.Infof("to backend %s %s", method, url)
	response, err := core.SendHttpRequest(ctx, method, url, body, headers, 10*time.Second)
	if err != nil {
		klog.Error(err)
		c.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "The back-end service is abnormal.", nil))
		return
	}
	// 设置响应头
	core.SetResHeaders(ctx, c, response.Header)

	klog.Debugf("body: %v", string(response.Body))
	klog.Info("return the response data.")
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), response.Body)
}

// NoAuthProxy 非认证代理
func NoAuthProxy(target string, path string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)
		klog.Debug("start proxy.")

		headers := core.SetReqHeaders(ctx, c)
		klog.Debugf("headers %+v", headers)
		method := string(c.Method())
		url := SetUrl(c, target, path)
		body, _ := c.Body()
		klog.Infof("to backend %s %s", method, url)

		// 发送http请求
		response, err := core.SendHttpRequest(ctx, method, url, body, headers, 10*time.Second)
		if err != nil {
			klog.Errorf("Send http err: %v", err)
			c.JSON(http.StatusInternalServerError,
				answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
			return
		}

		// 设置响应头
		core.SetResHeaders(ctx, c, response.Header)

		klog.Debugf("body: %v", string(response.Body))
		klog.Info("return the response data.")
		c.Data(response.StatusCode, response.Header.Get("Content-Type"), response.Body)
	}
}

// UiasAuthProxy Uias 认证代理
func UiasAuthProxy(target string, path string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)

		session := sessions.Default(c)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))

		if !isLogin { // 用户没有登录
			klog.Warn("user is not login. Please log in and try again")
			c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, ""))
			return
		}

		klog.Debug("The user has logged in.")
		StartProxy(ctx, c, session, target, path)
	}
}
