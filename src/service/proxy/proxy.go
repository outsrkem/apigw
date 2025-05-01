package proxy

import (
	"apigw/src/pkg/answer"
	"apigw/src/pkg/proxy"
	"apigw/src/slog"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
)

func tokenRenewal(token string) (string, error) {
	return token, nil
}

func StartProxy(ctx *app.RequestContext, session sessions.Session, host, rUrl string) {
	klog := slog.FromContext(ctx)
	body1, err := ctx.Body()
	if err != nil {
		klog.Errorf("body err: %v", err)
	}

	klog.Debugf("req body: %v", string(body1))
	payload := strings.NewReader(string(body1))
	klog.Debugf("req payload: %v", payload)

	method := string(ctx.Method())

	// 头部处理, 获取原有请求头并透传
	headers := make(map[string]string)
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	XSubjectToken, ok := session.Get("X-Subject-Token").(string)
	if ok {
		headers["X-Auth-Token"] = XSubjectToken       // 转换成功，可以使用
		session.Set("X-Subject-Token", XSubjectToken) // 保存一次session，防止过期
		if err := session.Save(); err != nil {
			klog.Errorf("session save error: %v", err)
		}
	}

	// TODO token 续签
	proxyPass := host + strings.TrimPrefix(string(ctx.URI().RequestURI()), rUrl) // 去除url前缀
	klog.Debugf("proxyPass %v", proxyPass)

	proxy, _ := proxy.NewProxy()
	res, _ := proxy.NewProxyRes(headers, method, proxyPass, payload)
	upstreamData, err := proxy.DoHttpV1(res)
	if err != nil {
		klog.Error(err)
		ctx.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "The back-end service is abnormal.", nil))
		return
	}
	defer func() {
		if upstreamData != nil {
			klog.Info("start close body.")
			klog.Debugf("upstreamData: %v", upstreamData)
			if err := upstreamData.Body.Close(); err != nil {
				klog.Errorf("Close request failed: %v", err)
			}
		}
	}()

	// 设置响应头
	for key, value := range upstreamData.Header {
		klog.Debugf("Header[%q] = %q", key, value)
		ctx.Response.Header.Set(key, strings.Join(value, ""))
	}

	// 处理响应体
	body, err := io.ReadAll(upstreamData.Body)
	if err != nil {
		klog.Errorf("read body err: %v", err)
		ctx.JSON(500, map[string]interface{}{})
	}

	klog.Debugf("body: %v", string(body))
	klog.Info("return the response data.")
	ctx.Data(upstreamData.StatusCode, string(ctx.Response.Header.ContentType()), body)
}

// ProxyUrl 反向代理的逻辑处理, 适用于浏览器用户登录后的接口调用
// 统一提权，向上游请求时，在请求头中添加token
func ProxyUrl(host string, rUrl string) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromContext(ctx)

		session := sessions.Default(ctx)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))

		if !isLogin { // 用户没有登录
			klog.Warn("user is not login. Please log in and try again")
			ctx.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, ""))
			return
		}

		klog.Debug("The user has logged in.")
		StartProxy(ctx, session, host, rUrl)
	}
}
