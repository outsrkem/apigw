package route

import (
	"fmt"
	"log"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app/server"

	"apigw/src/config"
	"apigw/src/mw"
	"apigw/src/user"
)

func RouteLocal(h *server.Hertz) {
	h.POST("/uias/v1/user/signin", user.UiasSignin)
	h.POST("/uias/v1/user/logout", user.UiasLogout)
}

func RouteProxy(h *server.Hertz, proxy *[]config.Proxy) {
	for _, apigw := range *proxy {
		for _, server := range apigw.Server {
			// 后端服务域名
			host := server.Location.Backend.Host
			// 目标url
			tUrl := server.Location.Backend.Url
			// 请求url
			rUrl := server.Location.Path
			fmt.Println(apigw.Name, ": ", server.Location.Path, "->", host+tUrl)
			//
			target, _ := url.Parse(host)
			log.Println(rUrl, target)

			h.HEAD(rUrl+"/*path", mw.ProxyUrl(host, rUrl))
			h.GET(rUrl+"/*path", mw.ProxyUrl(host, rUrl))
			h.POST(rUrl+"/*path", mw.ProxyUrl(host, rUrl))
			h.DELETE(rUrl+"/*path", mw.ProxyUrl(host, rUrl))
			h.PATCH(rUrl+"/*path", mw.ProxyUrl(host, rUrl))

		}
	}
}
