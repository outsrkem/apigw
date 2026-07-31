package handler

import (
	"apigw/src/cache"
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
		path := string(c.Request.Path())
		method := string(c.Request.Method())
		clientIP := getRealClientIP(c)
		klog.Infof("receive new request: %s %s %s", method, path, clientIP)
		klog.Debugf("raw request path:%s, request method:%s, real client ip:%s", path, method, clientIP)

		// Step 1: Route matching to locate the target api interface configuration
		klog.Debug("start route matching process")
		route, err := service.RouteSvc.MatchRoute(ctx, path, method)
		if err != nil || route == nil {
			klog.Errorf("route match failed %s %s", method, path)
			klog.Errorf("%v", err)
			c.JSON(404, answer.NewResMessage(answer.EcodeStatusNotFound, "The requested interface does not exist", nil))
			klog.Debug("request terminated, return 404 response")
			return
		}
		klog.Debugf("route match success, route id:%s, route status:%d, lc id:%s, auth type:%s", route.ID, route.Status, route.LcID, route.Auth)

		// Step 2. Status=0 means permanent disable, intercept request directly
		if route.Status == 0 {
			klog.Errorf("route is permanently disabled | route_id:%s | path:%s", route.ID, path)
			c.JSON(503, answer.NewResMessage(answer.EcodeUpstreamUnavailable, "The interface has been permanently disabled", nil))
			klog.Debug("request terminated, return 503 disabled response")
			return
		}
		klog.Debug("route status normal, continue forwarding process")

		// Step 3: Resolve real client IP, used for IP-Hash load balancing policy
		klog.Debugf("request client ip: %s, prepare load balance node selection", clientIP)

		// Step 4: Select available backend node via load balancing algorithm
		target, err := proxy.SelectTarget(route)
		if err != nil {
			klog.Errorf("select backend node failed | route_id:%s | client_ip:%s | err:%v", route.ID, clientIP, err)
			c.JSON(503, answer.NewResMessage(answer.EcodeUpstreamUnavailable, "Downstream service unavailable", nil))
			klog.Debug("request terminated, load balance node selection failed, return 503")
			return
		}
		klog.Infof("load balance select node success | backend_node:%s", target)
		klog.Debugf("selected upstream backend node address: %s", target)

		// Step 5: Fetch load channel configuration bound to current route
		channel := cache.GlobalCache.GetChannelByLCID(route.LcID)
		if channel == nil {
			klog.Errorf("load channel not found | lc_id:%s | route_id:%s", route.LcID, route.ID)
			c.JSON(503, answer.NewResMessage(answer.EcodeUpstreamUnavailable, "Load channel not found", nil))
			klog.Debug("request terminated, channel config missing, return 503")
			return
		}
		klog.Infof("get channel config success | channel_name:%s | raw_timeout:%dms", channel.Name, channel.Timeout)
		klog.Debugf("channel detailed config: name=%s, raw timeout=%dms", channel.Name, channel.Timeout)

		// Step 6: Normalize timeout value & create context with request timeout limit
		timeoutMs := GetSafeTimeout(channel.Timeout)
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
		reqDuration := time.Duration(timeoutMs) * time.Millisecond
		klog.Infof("request timeout normalized | final_timeout:%dms", timeoutMs)
		klog.Debugf("generate timeout context, effective timeout duration: %v", reqDuration)

		// Step 7: Build TLS configuration for HTTPS upstream connection
		// tlsCfg, err := proxy.BuildtlsCfg(channel.CaCert)
		// if err != nil {
		// 	klog.Errorf("build tls config fail | route_id:%s | lc_id:%s | err:%v", route.ID, route.LcID, err)
		// 	c.JSON(400, answer.NewResMessage(answer.EcodeParamInvalid, "CA certificate format error", nil))
		// 	klog.Debug("request terminated, tls config build failed, return 400")
		// 	return
		// }

		// Step 8: Rewrite request path according to route matching rule
		newUrl := RewritePath(path, route.ReqUri, route.BackendUri, route.Mode)
		klog.Infof("path rewrite complete | original_path:%s | rewrite_path:%s", path, newUrl)
		klog.Debugf("path rewrite params: reqUri=%s, backendUri=%s, rewriteMode=%s", route.ReqUri, route.BackendUri, route.Mode)

		// Step 9: Assemble complete request url with http/https schema
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
			klog.Debugf("append url query string: %s", query)
		}

		klog.Infof("forward target url: %s, request max timeout: %dms", url, timeoutMs)
		klog.Debugf("final assembled full forward url: %s", url)

		// Step 10: Extract raw request body and convert hertz request headers to map structure
		body, _ := c.Body()
		headers := proxy.SetReqHeaders(c)
		klog.Debugf("request body length:%d bytes | headers count:%d", len(body), len(headers))

		// Step 11: build cfg
		cfg := proxy.GetProxyRequest()
		klog.Debug("obtain RequestArgs instance from object pool")
		cfg.Method = method
		cfg.Protocol = route.Protocol
		cfg.LcCaID = channel.CaCert
		cfg.Url = url
		cfg.Body = body
		cfg.Headers = headers
		cfg.ReqTimeout = reqDuration
		// cfg.TlsCfg = tlsCfg
		klog.Debug("finish filling all RequestArgs forwarding parameters")
		defer func() {
			proxy.PutProxyRequest(cfg)
			klog.Debug("RequestArgs recycled back to object pool after forwarding")
		}()

		// Distribute forwarding logic by authentication type
		authType := strings.ToUpper(route.Auth)
		klog.Debugf("resolved route authentication type: %s", authType)
		switch authType {
		case "NONE":
			klog.Infof("use NONE authentication proxy mode")
			klog.Debug("enter NoAuthProxy forwarding function")
			proxy.NoAuthProxy(reqCtx, c, route.LcID, cfg)
		case "UIAS":
			klog.Infof("use UIAS authentication proxy mode")
			klog.Debug("enter UiasAuthProxy forwarding function")
			proxy.UiasAuthProxy(reqCtx, c, route.LcID, cfg)
		case "TOKEN":
			klog.Infof("use TOKEN authentication proxy mode")
			klog.Debug("enter TokenAuthProxy forwarding function")
			proxy.TokenAuthProxy(reqCtx, c, route.LcID, cfg)
		default:
			klog.Errorf("Incorrect interface authentication configuration type | route_id:%s | current_auth:%s, support:[NONE,UIAS,TOKEN]", route.ID, authType)
			c.JSON(400, answer.NewResMessage(answer.EcodeParamInvalid, "Incorrect interface authentication configuration type", nil))
			klog.Debug("request terminated, illegal auth type, return 400")
		}
		klog.Debug("ProxyDispatch overall forwarding flow finished")
	}
}
