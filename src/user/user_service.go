package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"

	"apigw/src/mw"
	"apigw/src/pkg/proxy"
)

// @router /api/uias/v1/user/signin [POST]

func UiasSignin(host string, url string) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		session := sessions.Default(ctx)
		islogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("islogin")))
		if islogin {
			mw.ResponseBody(ctx, http.StatusOK, "User logged in.")
			return
		}

		headers := make(map[string]string)
		// 获取原有请求头并透传
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})
		headers["Content-Type"] = "application/json;charset=utf-8"

		method := string(ctx.Method())
		body1, _ := ctx.Body()
		payload := strings.NewReader(string(body1))

		//	发送http请求
		proxy_pass := host + url
		proxy, _ := proxy.NewProxy()
		res, _ := proxy.NewProxyRes(headers, method, proxy_pass, payload)
		answer, err := proxy.DoHttpV1(res)
		if err != nil {
			log.Println(err)
			mw.ResponseBody(ctx, http.StatusInternalServerError, "The back-end service is abnormal.")
			return
		}
		// 不知道defer放这里会不会有问题
		defer answer.Body.Close()
		ctx.Response.Header.GetHeaders()
		for key, value := range ctx.Response.Header.GetHeaders() {
			fmt.Println(key, value)
		}
		// 处理响应头
		request_id := answer.Header.Get("X-Request-Id")
		ctx.Response.Header.Set("X-Request-Id", request_id)

		// 处理响应体
		body, _ := io.ReadAll(answer.Body)
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(string(body)), &result); err != nil {
			log.Println("json Unmarshal error: ", err)
		}

		// 获取响应状态码
		sCode := answer.StatusCode
		if sCode == 200 {
			log.Println("Login successfully.")
			// 登录成功后保存token
			XSubjectToken := answer.Header.Get("X-Subject-Token")
			session.Set("X-Subject-Token", XSubjectToken)
			session.Set("islogin", true)
			_ = session.Save()
		}
		mw.ResponseBody(ctx, sCode, result)
	}
}

// @router /api/uias/v1/user/logout [POST]
func UiasLogout(c context.Context, ctx *app.RequestContext) {
	session := sessions.Default(ctx)
	islogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("islogin")))
	if islogin {
		log.Println("logout successfully.")
		// session.Options(sessions.Options{
		// 	MaxAge: 0,
		// })
		session.Clear()
		session.Save()
		mw.ResponseBody(ctx, http.StatusOK, nil)
		return
	}
	mw.ResponseBody(ctx, http.StatusBadRequest, "User not logged in.")
}
