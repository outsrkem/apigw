package route

import (
	"log"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app/server"

	"apigw/src/config"
	"apigw/src/mw"
	"apigw/src/user"
)

func RouteLocal(h *server.Hertz, auth *config.Auth) {
	host := auth.Backend.Host
	url := "/internal/v1/uias/user/signin"
	log.Println("auth : /uias/v1/user/signin -> " + host + url)

	h.POST("/api/uias/v1/user/signin", user.UiasSignin(host, url))
	h.POST("/api/uias/v1/user/logout", user.UiasLogout)
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
			log.Println(apigw.Name, ": ", server.Location.Path, "->", host+tUrl)
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
