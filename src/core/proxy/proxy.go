package proxy

import (
	"apigw/src/pkg/answer"
	"apigw/src/pkg/consts"
	"apigw/src/pkg/redisx"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/sessions"
	"github.com/sirupsen/logrus"
)

type RequestArgs struct {
	Method     string
	Protocol   string
	LcCaID     string
	Url        string
	Body       []byte
	Headers    map[string]string
	ReqTimeout time.Duration
}

// proxyRequestPool object pool for reusing RequestArgs instances to reduce GC pressure
var proxyRequestPool = sync.Pool{
	New: func() any {
		return &RequestArgs{}
	},
}

var AuthBackendUrl = ""

// doRenewToken implements token refresh logic
// return newToken: new token string, newExpTs: X‑Token‑ExpireAt timestamp in milliseconds
func doRenewToken(oldToken string) (newToken string, newExpTs int64, err error) {
	ctx := context.Background()
	const url = "/v1/uias/auth/token/refresh"
	headers := map[string]string{consts.HeaderXAuthToken: oldToken}
	response, err := HeProxy(ctx, "POST", "HTTP", "", AuthBackendUrl+url, nil, headers)
	if err != nil {
		return "", 0, err
	}

	XTokenExpireAt := response.Header.Get(consts.HeaderXTokenExpireAt)
	expTs, err := utils.Rfc1123ToMilli(XTokenExpireAt)
	if err != nil {
		return "", 0, err
	}

	XSubjectToken := response.Header.Get(consts.HeaderXSubjectToken)
	return XSubjectToken, expTs, nil
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
	proxyRequestPool.Put(req)
}

// NoAuthProxy proxy without login authentication verification
func NoAuthProxy(ctx context.Context, c *app.RequestContext, lcID string, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	klog.Debugf("NoAuthProxy start forward, method: %s, target url: %s, body length: %d, timeout: %v",
		req.Method, req.Url, len(req.Body), req.ReqTimeout)

	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("send upstream http request failed, err: %v", err)
		c.JSON(http.StatusInternalServerError,
			answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
		return
	}

	klog.Debug("NoAuthProxy finished writing upstream headers to client response")

	c.Status(response.StatusCode())
	response.Header.VisitAll(func(key, value []byte) {
		c.Response.Header.Set(string(key), string(value))
	})

	c.Response.SetBody(response.Body())
	klog.Debug("NoAuthProxy response write to client complete")
}

// tokenRenewalAsync pre‑refresh token executed in async goroutine;
// current request keeps using old token, new token takes effect on subsequent requests
func tokenRenewalAsync(ctx context.Context, klog *logrus.Entry, sid string) {
	klog.Infof("token renewal async sid: %s", sid)
	mgr := redisx.Manager
	nowMs := time.Now().UnixMilli()

	ok, err := mgr.AgwSession.TryAcquireRefreshingLock(ctx, sid, nowMs, redisx.RefreshingLockTimeoutMs)
	if err != nil {
		klog.Errorf("tokenRenewalAsync try acquire lock err, sid=%s err:%v", sid, err)
		return
	}
	if !ok {
		klog.Debugf("tokenRenewalAsync already in refreshing or session gone, sid=%s", sid)
		return
	}

	defer func() {
		clearCtx := context.WithoutCancel(ctx)
		_ = mgr.AgwSession.ClearRefreshingLock(clearCtx, sid)
	}()

	sess, err := mgr.AgwSession.Get(ctx, sid)
	if err != nil {
		klog.Errorf("tokenRenewalAsync get session from redis failed, sid=%s err:%v", sid, err)
		return
	}

	oldToken := sess.XSubjectToken
	newToken, newExpTs, err := doRenewToken(oldToken)
	if err != nil {
		klog.Errorf("doRenewToken failed sid=%s err:%v", sid, err)
		return
	}
	if newExpTs <= nowMs {
		klog.Errorf("renew token invalid expire time sid=%s newExpTs=%d", sid, newExpTs)
		return
	}

	newRefreshPointTs := utils.CalcOneThirdTimePoint(newExpTs)
	ttlMs := newExpTs - nowMs
	if ttlMs < 0 {
		klog.Errorf("renewed token already expired sid=%s newExpTs=%d", sid, newExpTs)
		return
	}
	// redis expiration = token expiration time plus 30 minutes margin
	redisTTL := time.Duration(ttlMs)*time.Millisecond + consts.AgwSessionTTLMargin

	err = mgr.AgwSession.UpdateToken(ctx, sid, newToken, newRefreshPointTs, newExpTs, redisTTL)
	if err != nil {
		klog.Errorf("tokenRenewalAsync update redis session failed sid=%s err:%v", sid, err)
		return
	}
	klog.Infof("token renew success sid=%s, newRefreshPointTs=%d, newExpTs=%d", sid, newRefreshPointTs, newExpTs)
}

// UiasAuthProxy proxy with UIAS session login authentication verification
func UiasAuthProxy(ctx context.Context, c *app.RequestContext, lcID string, req *RequestArgs) {
	klog := slog.FromCtx(ctx)
	klog.Debugf("UiasAuthProxy enter authentication check, method: %s, target url: %s", req.Method, req.Url)

	session := sessions.Default(c)
	isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get(consts.SessionKeyIsLogin)))
	if !isLogin {
		klog.Warn("user is not login. Please log in and try again")
		c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, nil))
		return
	}

	sid, ok := session.Get(consts.RedisSessionKey).(string)
	if !ok {
		klog.Warnf("sid not found in session, %s", session.ID())
		c.JSON(401, answer.ResBody(answer.EcodeNotLogIn, "user not logged in.", nil))
		return
	}
	klog.Infof("sid: [%s]", sid)

	redisReadCtx, redisReadCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer redisReadCancel()
	sess, err := redisx.Manager.AgwSession.Get(redisReadCtx, sid)
	if err != nil {
		klog.Errorf("read redis session failed %s %v", sid, err)
		c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, nil))
		return
	}
	klog.Debugf("session loaded, sid=%s IsLogin=%t TokenExpireTs=%d RefreshPointTs=%d IdleTimeoutMs=%d LastAccessTs=%d",
		sid, sess.IsLogin, sess.TokenExpireTs, sess.RefreshPointTs, sess.IdleTimeoutMs, sess.LastAccessTs)

	if !sess.IsLogin {
		klog.Warnf("redis session isLogin false, %s", sid)
		c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, nil))
		return
	}

	nowMs := time.Now().UnixMilli()
	// Session idle timeout logout
	if sess.IdleTimeoutMs > 0 {
		idleDuration := nowMs - sess.LastAccessTs // Calculate elapsed milliseconds since last user request
		if idleDuration > sess.IdleTimeoutMs {    // Check whether idle duration exceeds configured idle‑timeout
			klog.Warnf("session idle timeout sid=%s idle=%d > idleTimeout=%d", sid, idleDuration, sess.IdleTimeoutMs)
			if err := redisx.Manager.AgwSession.Delete(context.Background(), sid); err != nil {
				klog.Errorf("idle timeout delete redis session failed, sid=%s err=%v", sid, err)
			}

			opt := sessions.Options{MaxAge: 0, Path: "/", HttpOnly: true}
			session.Options(opt)
			session.Clear()
			if err := session.Save(); err != nil {
				klog.Errorf("idle timeout clear session cookie save failed, sid=%s err=%v", sid, err)
			}

			c.JSON(http.StatusUnauthorized, answer.ResBody(answer.EcodeNotLogIn, nil, nil))
			return
		}
	}

	// Cookie sliding renewal
	nowSec := nowMs / 1000                                                // Current unix timestamp in seconds
	lastRefreshSec := sess.LastCookieRefreshTs / 1000                     // Last refresh timestamp in seconds
	cookieRemainSec := consts.CookieMaxAgeSec - (nowSec - lastRefreshSec) // Calculate remaining cookie lifetime in seconds

	klog.Debugf("cookie sliding renewal check, sid=%s lastCookieRefreshTs=%d nowSec=%d cookieRemainSec=%d threshold=%d",
		sid, sess.LastCookieRefreshTs, nowSec, cookieRemainSec, consts.CookieRenewThresholdSec)

	if cookieRemainSec < consts.CookieRenewThresholdSec && cookieRemainSec > 0 {
		newRefreshMs := nowMs
		go func(sid string, tsMs int64) {
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("UpdateLastCookieRefreshTs panic, sid=%s err=%v", sid, r)
				}
			}()
			_ = redisx.Manager.AgwSession.UpdateLastCookieRefreshTs(context.Background(), sid, tsMs)
		}(sid, newRefreshMs)

		// Must set a value to make session.Save take effect; this value has no business meaning
		session.Set("last_cookie_refresh_ts", newRefreshMs)
		if err := session.Save(); err != nil {
			klog.Errorf("session save error: %v", err)
		}

		klog.Infof("cookie sliding renewal trigger, sid=%s oldLastCookieRefreshTs=%d cookieRemainSec=%d newRefreshMs=%d",
			sid, sess.LastCookieRefreshTs, cookieRemainSec, newRefreshMs)
	}

	// Async update last‑access timestamp and slide redis session ttl
	redisSessionTTL := time.Duration(consts.CookieMaxAgeSec+600) * time.Second // Redis ttl adds extra 10min margin compared with cookie
	go func(sid string, ts int64, ttl time.Duration) {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("UpdateLastAccessAndExpire panic recover sid=%s, panic=%v", sid, r)
			}
		}()
		_ = redisx.Manager.AgwSession.UpdateLastAccessAndExpire(context.Background(), sid, ts, ttl)
	}(sid, nowMs, redisSessionTTL)

	// Trigger async token refresh goroutine if refresh time point is reached
	if nowMs >= sess.RefreshPointTs {
		go func(sid string) {
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("tokenRenewalAsync panic recover sid=%s, panic=%v", sid, r)
				}
			}()
			tokenRenewalAsync(context.Background(), klog, sid)
		}(sid)
	}

	// Inject token into request header for upstream forwarding
	klog.Debugf("UiasAuthProxy start forward request, body length: %d, timeout: %v", len(req.Body), req.ReqTimeout)

	req.Headers[consts.HeaderXAuthToken] = sess.XSubjectToken
	response, err := HeProxy(ctx, req.Method, req.Protocol, req.LcCaID, req.Url, req.Body, req.Headers)
	defer protocol.ReleaseResponse(response)
	if err != nil {
		klog.Errorf("Send http err: %v", err)
		c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeReadUpstreamDataError, "Internal Server Error", nil))
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
	klog.Debugf("TokenAuthProxy start forward, method: %s, target url: %s, body length: %d, timeout: %v",
		req.Method, req.Url, len(req.Body), req.ReqTimeout)

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
