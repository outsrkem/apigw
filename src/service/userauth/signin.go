package userauth

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

// UiasSignin 登录
func UiasSignin(host string, url string) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromContext(ctx)
		session := sessions.Default(ctx)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))
		if isLogin {
			ctx.JSON(200, answer.NewResMessage(answer.EcodeOkay, "User logged in.", nil))
			return
		}

		headers := make(map[string]string)
		// 获取原有请求头并透传
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		method := string(ctx.Method())
		body1, _ := ctx.Body()
		payload := strings.NewReader(string(body1))

		// 发送http请求
		proxyPass := host + url
		_proxy, _ := proxy.NewProxy()
		res, _ := _proxy.NewProxyRes(headers, method, proxyPass, payload)

		klog.Debugf("headers: %v", headers)
		upstreamData, err := _proxy.DoHttpV1(res)
		if err != nil {
			klog.Errorf("Sending request error: %v.", err)
			ctx.JSON(500, answer.NewResMessage(answer.EcodeSendingRequest, "Sending request error.", nil))
			return
		}
		defer func() {
			if upstreamData != nil {
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
			ctx.JSON(http.StatusInternalServerError,
				answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
			return
		}

		if upstreamData.StatusCode == 200 {
			// 登录成功后保存token
			klog.Info("log in success.")
			XSubjectToken := upstreamData.Header.Get("X-Subject-Token")
			session.Set("X-Subject-Token", XSubjectToken)
			session.Set("isLogin", true)
			if err = session.Save(); err != nil {
				klog.Errorf("session save err: %v", err)
				ctx.JSON(http.StatusInternalServerError,
					answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}
			klog.Info("log in success.")
		}

		// 无论成功或失败都透传上游的响应
		ctx.Data(upstreamData.StatusCode, string(ctx.Response.Header.ContentType()), body)
	}
}
