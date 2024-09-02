package mw

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

	"apigw/src/pkg/proxy"
)

func tokenRenewal(token string) (string, error) {
	return token, nil
}

// 反向代理的逻辑处理
// 适用于浏览器用户登录后的接口调用
// 统一提权，向上游请求时，在请求头中添加token
func ProxyUrl(host string, rUrl string) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		session := sessions.Default(ctx)
		islogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("islogin")))
		if islogin {
			body1, _ := ctx.Body()
			payload := strings.NewReader(string(body1))
			method := string(ctx.Method())

			// 头部处理
			headers := make(map[string]string)
			XSubjectToken, ok := session.Get("X-Subject-Token").(string)
			if ok {
				// 转换成功，可以使用
				headers["X-Auth-Token"] = XSubjectToken
				// 保存一次session，防止过期
				session.Set("X-Subject-Token", XSubjectToken)
				_ = session.Save()
			}
			// token 续签
			if 1 == 2 {
				newtoken, err := tokenRenewal(XSubjectToken)
				if err != nil {
					log.Println()
				}
				session.Set("X-Subject-Token", newtoken)
				_ = session.Save()
			}

			// https://www.cloudwego.io/zh/docs/hertz/tutorials/basic-feature/context/request/#uri-%E5%AF%B9%E8%B1%A1
			// 去除url前缀
			proxy_pass := host + strings.TrimPrefix(string(ctx.URI().RequestURI()), rUrl)
			proxy, _ := proxy.NewProxy()
			res, _ := proxy.NewProxyRes(headers, method, proxy_pass, payload)
			answer, err := proxy.DoHttpV1(res)
			if err != nil {
				log.Println(err)
				ResponseBody(ctx, http.StatusInternalServerError, "The back-end service is abnormal.")
				return
			}

			// 不知道defer放这里会不会有问题
			defer answer.Body.Close()

			//设置响应头
			for key, value := range answer.Header {
				ctx.Response.Header.Set(key, strings.Join(value, ""))
			}

			// 处理响应体
			body, _ := io.ReadAll(answer.Body)
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(string(body)), &result); err != nil {
				log.Println("json Unmarshal error: ", err)
			}
			// 设置响应状态码
			sCode := answer.StatusCode
			ctx.JSON(sCode, result)
			return
		}
		// 用户没有登录
		ResponseBody(ctx, http.StatusUnauthorized, nil)
	}
}
