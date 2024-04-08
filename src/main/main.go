package main

import (
	"apigw/src/config"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	// 解析配置文件
	cfgApigw := config.InitConfig()

	for _, apigw := range cfgApigw.Apigw {
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
			proxy := httputil.NewSingleHostReverseProxy(target)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println("request url:", r.URL.Path)
				// 移除前缀
				newPath := strings.TrimPrefix(r.URL.Path, rUrl)
				if tUrl == "/" {
					r.URL.Path = newPath
				} else {
					r.URL.Path = tUrl + newPath
				}
				if r.URL.Path == "/v1/uias/user/signin" {
					r.URL.Path = "/internal/v1/uias/user/signin"
				}
				if r.URL.Path == "/v1/uias/user/register" {
					r.URL.Path = "/internal/v1/uias/user/register"
				}
				log.Println("target url:", r.URL.Path)
				log.Println("target host:", host)
				proxy.ServeHTTP(w, r)
			})
			http.Handle(rUrl+"/*", handler)
		}
	}

	http.ListenAndServe(":8080", nil)
}
