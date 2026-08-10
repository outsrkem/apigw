// src/handler/handler.go
package handler

import (
	"apigw/src/core/proxy"
	"apigw/src/core/service"
	"apigw/src/pkg/answer"
	"apigw/src/slog"
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ProxyDispatch returns hertz middleware handler function for unified proxy forwarding scheduling
func ProxyDispatch() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		host := GetMatchHost(c)
		path := string(c.Request.Path())
		method := string(c.Request.Method())
		clientIP := getRealClientIP(c)
		klog.Infof("receive new request: %s %s %s %s", host, method, path, clientIP)
		klog.Debugf("raw request path:%s, request method:%s, real client ip:%s", path, method, clientIP)

		// Step 1: Route matching to locate the target api interface configuration
		klog.Debug("start route matching process")
		route, err := service.RouteSvc.MatchRoute(ctx, host, path, method)
		if err != nil || route == nil {
			klog.Errorf("route match failed %s %s", method, path)
			klog.Errorf("%v", err)
			c.JSON(404, answer.NewResMessage(answer.EcodeStatusNotFound, "api does not exist", nil))
			return
		}
		klog.Debugf("route match success, route id:%s,  lc id:%s, auth type:%s", route.ID, route.LcID, route.Auth)

		// Step 4: 选择后端节点
		target, err := proxy.SelectTarget(route.LcID)
		if err != nil {
			klog.Errorf("select backend node failed | route_id:%s | client_ip:%s | err:%v", route.ID, clientIP, err)
			c.JSON(503, answer.NewResMessage(answer.EcodeUpstreamUnavailable, "Downstream service unavailable", nil))
			return
		}
		klog.Infof("load balance select node success | backend_node:%s", target)

		// Step 5: 获取绑定到当前路线的负载通道配置
		channel := service.RouteSvc.GetChannel(route.LcID)
		if channel == nil {
			klog.Errorf("load channel not found | lc_id:%s | route_id:%s", route.LcID, route.ID)
			c.JSON(503, answer.NewResMessage(answer.EcodeUpstreamUnavailable, "Load channel not found", nil))
			return
		}
		klog.Infof("get channel config success | channel_name:%s | raw_timeout:%dms | channel_status:%d", channel.Name, channel.Timeout, channel.Status)

		// Step 6: Normalize timeout value & create context with request timeout limit
		timeoutMs := GetSafeTimeout(channel.Timeout)
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
		reqDuration := time.Duration(timeoutMs) * time.Millisecond
		klog.Infof("request timeout normalized | final_timeout:%dms", timeoutMs)

		// Step 7: Rewrite request path according to route matching rule
		newUrl := RewritePath(path, route.ReqUri, route.BackendUri, route.Mode)
		klog.Infof("path rewrite complete | original_path:%s | rewrite_path:%s", path, newUrl)
		klog.Debugf("path rewrite params: reqUri=%s, backendUri=%s, rewriteMode=%s", route.ReqUri, route.BackendUri, route.Mode)

		// Step 8: Assemble complete request url with http/https schema
		var schema string
		switch strings.ToUpper(route.Protocol) {
		case "HTTPS":
			schema = "https://"
			klog.Debug("protocol identified as HTTPS")
		default:
			schema = "http://"
			klog.Debug("protocol identified as HTTP")
		}
		url := schema + target + newUrl
		query := string(c.URI().QueryString())
		if query != "" {
			url += "?" + query
		}

		klog.Infof("forward target url: %s, request max timeout: %dms", url, timeoutMs)

		// Step 9: Extract raw request body and convert hertz request headers to map structure
		body, _ := c.Body()
		headers := proxy.SetReqHeaders(c)

		// Step 10: build proxy request cfg
		cfg := proxy.GetProxyRequest()
		klog.Debug("obtain RequestArgs instance from object pool")
		cfg.Method = method
		cfg.Protocol = route.Protocol
		cfg.LcCaID = channel.CaCert
		cfg.Url = url
		cfg.Body = body
		cfg.Headers = headers
		cfg.ReqTimeout = reqDuration
		defer func() {
			proxy.PutProxyRequest(cfg)
		}()

		// Step 11: Distribute forwarding logic by authentication type
		authType := strings.ToUpper(route.Auth)
		klog.Debugf("resolved route authentication type: %s", authType)
		switch authType {
		case "NONE":
			proxy.NoAuthProxy(reqCtx, c, route.LcID, cfg)
		case "UIAS":
			proxy.UiasAuthProxy(reqCtx, c, route.LcID, cfg)
		case "TOKEN":
			proxy.TokenAuthProxy(reqCtx, c, route.LcID, cfg)
		default:
			klog.Errorf("Incorrect interface authentication configuration type | route_id:%s | current_auth:%s, support:[NONE,UIAS,TOKEN]", route.ID, authType)
			c.JSON(400, answer.NewResMessage(answer.EcodeParamInvalid, "Incorrect interface authentication configuration type", nil))
		}
	}
}
