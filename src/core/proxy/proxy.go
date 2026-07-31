package proxy

import (
	"apigw/src/pkg/answer"
	"apigw/src/slog"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
)

type RequestArgs struct {
	Method     string
	Protocol   string
	LcCaID     string
	Url        string
	Body       []byte
	Headers    map[string]string
	ReqTimeout time.Duration
	TlsCfg     *tls.Config
}

// proxyRequestPool object pool for reusing RequestArgs instances to reduce GC pressure
var proxyRequestPool = sync.Pool{
	New: func() interface{} {
		return &RequestArgs{}
	},
}

// GetProxyRequest fetches an available RequestArgs instance from the object pool.
func GetProxyRequest() *RequestArgs {
	return proxyRequestPool.Get().(*RequestArgs)
}

// PutProxyRequest recycles RequestArgs back to the object pool.
func PutProxyRequest(req *RequestArgs) {
	if req == nil {
		return
	}
	req.Method = ""
	req.Protocol = ""
	req.LcCaID = ""
	req.Url = ""
	req.Body = nil
	req.Headers = nil
	req.ReqTimeout = 0
	req.TlsCfg = nil
	proxyRequestPool.Put(req)
}

// NoAuthProxy proxy without login authentication verification
func NoAuthProxy(ctx context.Context, c *app.RequestContext, lcID string, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	klog.Debugf("NoAuthProxy start forward, method: %s, target url: %s, body length: %d, timeout: %v, enableTLS: %t",
		req.Method, req.Url, len(req.Body), req.ReqTimeout, req.TlsCfg != nil)

	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("send upstream http request failed, err: %v", err)
		c.JSON(http.StatusInternalServerError,
			answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
		return
	}
	klog.Debug("NoAuthProxy get upstream response successfully")

	klog.Debug("NoAuthProxy finished writing upstream headers to client response")

	c.Status(response.StatusCode())
	response.Header.VisitAll(func(key, value []byte) {
		c.Response.Header.Set(string(key), string(value))
	})

	c.Response.SetBody(response.Body())
	klog.Debug("NoAuthProxy response write to client complete")
}

// UiasAuthProxy proxy with UIAS session login authentication verification
func UiasAuthProxy(ctx context.Context, c *app.RequestContext, lcID string, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	klog.Debugf("UiasAuthProxy enter authentication check, method: %s, target url: %s", req.Method, req.Url)

	session := sessions.Default(c)
	loginRaw := session.Get("isLogin")
	klog.Debugf("read session isLogin raw value: %v", loginRaw)
	isLogin, _ := strconv.ParseBool(fmt.Sprint(loginRaw))
	klog.Debugf("parsed user login status: %t", isLogin)

	if !isLogin {
		klog.Warn("user is not login. Please log in and try again")
		c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, nil))
		klog.Debug("UiasAuthProxy reject request: user not logged in, return 401")
		return
	}
	klog.Debug("UiasAuthProxy login authentication passed")

	// Token renewal logic
	if XSubjectToken, ok := session.Get("X-Subject-Token").(string); ok {
		req.Headers["X-Auth-Token"] = XSubjectToken
		session.Set("X-Subject-Token", XSubjectToken)
		if err := session.Save(); err != nil {
			klog.Errorf("session save error: %v", err)
		} else {
			klog.Debug("UiasAuthProxy refresh session token and save session success")
		}
	} else {
		klog.Debug("X-Subject-Token not found in current session")
	}

	klog.Debugf("UiasAuthProxy start forward request, body length: %d, timeout: %v, enableTLS: %t",
		len(req.Body), req.ReqTimeout, req.TlsCfg != nil)

	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("Send http err: %v", err)
		c.JSON(http.StatusInternalServerError,
			answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
		return
	}

	c.Status(response.StatusCode())
	response.Header.VisitAll(func(key, value []byte) {
		c.Response.Header.Set(string(key), string(value))
	})

	c.Response.SetBody(response.Body())
}

// TokenAuthProxy proxy with external token authentication verification
func TokenAuthProxy(ctx context.Context, c *app.RequestContext, lcID string, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	klog.Debugf("TokenAuthProxy start forward, method: %s, target url: %s, body length: %d, timeout: %v, enableTLS: %t",
		req.Method, req.Url, len(req.Body), req.ReqTimeout, req.TlsCfg != nil)

	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("Send http err: %v", err)
		c.JSON(http.StatusInternalServerError,
			answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
		return
	}

	c.Status(response.StatusCode())
	response.Header.VisitAll(func(key, value []byte) {
		c.Response.Header.Set(string(key), string(value))
	})

	c.Response.SetBody(response.Body())
}

func InternalHttpProxy(ctx context.Context, c *app.RequestContext, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	method := req.Method
	targetUrl := req.Url
	reqBodyLen := len(req.Body)
	timeout := req.ReqTimeout
	enableTLS := req.TlsCfg != nil

	klog.Debugf("InternalHttpProxy start forward | method=%s target=%s reqBodyLen=%d timeout=%v enableTLS=%t",
		method, targetUrl, reqBodyLen, timeout, enableTLS)

	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("InternalHttpProxy send upstream request failed | method=%s target=%s err=%v", method, targetUrl, err)
		c.JSON(http.StatusInternalServerError,
			answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
		return
	}
	c.Status(response.StatusCode())
	response.Header.VisitAll(func(key, value []byte) {
		c.Response.Header.Set(string(key), string(value))
	})

	c.Response.SetBody(response.Body())
}
