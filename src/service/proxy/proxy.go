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

func StartProxy(ctx context.Context, c *app.RequestContext, session sessions.Session, host, path string) {
	klog := slog.FromContext(c)
	headers := core.SetReqHeaders(ctx, c) // 头部处理, 获取原有请求头并透传
	method := string(c.Method())
	body, err := c.Body()
	if err != nil {
		klog.Errorf("body err: %v", err)
	}
	klog.Debugf("req body: %v", string(body))

	XSubjectToken, ok := session.Get("X-Subject-Token").(string)
	if ok {
		headers["X-Auth-Token"] = XSubjectToken       // 转换成功，可以使用
		session.Set("X-Subject-Token", XSubjectToken) // 保存一次session，防止过期
		if err := session.Save(); err != nil {
			klog.Errorf("session save error: %v", err)
		}
	}

	// TODO token 续签
	url := host + strings.TrimPrefix(string(c.URI().RequestURI()), path) // 去除url前缀
	klog.Debugf("proxyPass %v", url)

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

// ProxyUrl 反向代理的逻辑处理, 适用于浏览器用户登录后的接口调用
// 统一提权，向上游请求时，在请求头中添加token
func ProxyUrl(host string, rUrl string) func(ctx context.Context, c *app.RequestContext) {
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
		StartProxy(ctx, c, session, host, rUrl)
	}
}

// NoAuthProxy 非认证代理
func NoAuthProxy(host string, path string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)
		klog.Debug("start proxy.")

		headers := core.SetReqHeaders(ctx, c)
		klog.Debugf("headers %+v", headers)
		method := string(c.Method())
		url := host + path
		body, _ := c.Body()

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

// UiasAuthProxy Uias认证代理
func UiasAuthProxy(host string, path string) func(ctx context.Context, c *app.RequestContext) {
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
		StartProxy(ctx, c, session, host, path)
	}
}
