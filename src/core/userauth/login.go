package userauth

import (
	"apigw/src/core/proxy"
	"apigw/src/pkg/answer"
	"apigw/src/pkg/consts"
	"apigw/src/pkg/redisx"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/sessions"
	"github.com/sirupsen/logrus"
)

// Blacklist of response headers that should not be passed through to client
var blockRespHeaders = map[string]bool{
	consts.HeaderXSubjectToken:       true,
	consts.HeaderXTokenExpireAt:      true,
	consts.HeaderXSessionIdleTimeout: true,
}

// defaultIdleMin default session idle timeout in minutes
const defaultIdleMin int64 = 30

// parseXSessionIdleTimeout parse response header X‑Session‑Idle‑Timeout, value unit: minute
// idleHeaderVal: raw http header value
// sid: current session id for logging
// fallbackMin: fallback minutes when parse failed or header is empty
// return: idleTimeoutMs idle timeout in milliseconds
func parseXSessionIdleTimeout(idleHeaderVal string, sid string, fallbackMin int64, klog *logrus.Entry) int64 {
	var idleTimeoutMs int64
	if idleHeaderVal != "" {
		minVal, err := strconv.ParseInt(idleHeaderVal, 10, 64)
		if err != nil {
			klog.Warnf("parse %s failed, val=%s, err=%v, sid=%s, fallback to %dmin",
				consts.HeaderXSessionIdleTimeout, idleHeaderVal, err, sid, fallbackMin)
			idleTimeoutMs = fallbackMin * 60 * 1000
		} else if minVal > 0 {
			idleTimeoutMs = minVal * 60 * 1000
			klog.Debugf("%s use header value, minVal=%d, idleTimeoutMs=%d, sid=%s",
				consts.HeaderXSessionIdleTimeout, minVal, idleTimeoutMs, sid)
		} else {
			klog.Warnf("%s invalid value, minVal=%d, sid=%s, fallback to %dmin",
				consts.HeaderXSessionIdleTimeout, minVal, sid, fallbackMin)
			idleTimeoutMs = fallbackMin * 60 * 1000
		}
	} else {
		klog.Debugf("%s header empty, sid=%s, fallback to %dmin",
			consts.HeaderXSessionIdleTimeout, sid, fallbackMin)
		idleTimeoutMs = fallbackMin * 60 * 1000
	}
	klog.Debugf("final resolved idleTimeoutMs=%d sid=%s", idleTimeoutMs, sid)
	return idleTimeoutMs
}

// UiasSignIn handles login request
func UiasSignIn(host string, url string) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		session := sessions.Default(c)

		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))
		if isLogin {
			c.JSON(200, answer.NewResMessage(answer.EcodeOkay, "User logged in.", nil))
			return
		}

		headers := proxy.SetReqHeaders(c)
		method := string(c.Method())
		body, _ := c.Body()

		// Send http request to upstream login endpoint
		proxyPass := host + url

		response, err := proxy.HeProxy(ctx, method, "HTTP", "", proxyPass, body, headers)
		defer protocol.ReleaseResponse(response)
		if err != nil {
			klog.Error(err)
			c.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "The back-end service is abnormal.", nil))
			return
		}

		// Persist token after successful login
		if response.StatusCode() == 200 {
			sid := utils.CreateUuid()
			klog.Debugf("agw session id %s", sid)

			XSubjectToken := response.Header.Get(consts.HeaderXSubjectToken)
			if XSubjectToken == "" {
				klog.Errorf("%s is empty, sid=%s", consts.HeaderXSubjectToken, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}

			XTokenExpireAt := response.Header.Get(consts.HeaderXTokenExpireAt)
			if XTokenExpireAt == "" {
				klog.Errorf("%s header empty, sid=%s", consts.HeaderXTokenExpireAt, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}

			klog.Debugf("%s: %q", consts.HeaderXTokenExpireAt, XTokenExpireAt)
			tokenExpireAtMs, err := utils.Rfc1123ToMilli(XTokenExpireAt)
			if err != nil {
				klog.Errorf("parse %s failed, val=%s err=%v sid=%s", consts.HeaderXTokenExpireAt, XTokenExpireAt, err, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}

			nowMs := time.Now().UnixMilli()
			refreshPointTs := utils.CalcOneThirdTimePoint(tokenExpireAtMs)

			ttlMs := tokenExpireAtMs - nowMs
			if ttlMs < 0 {
				klog.Errorf("token already expired, tokenExpireAtMs=%d nowMs=%d sid=%s", tokenExpireAtMs, nowMs, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "token expired", nil))
				return
			}
			sttl := time.Duration(ttlMs)*time.Millisecond + consts.AgwSessionTTLMargin

			// Parse session idle timeout: X‑Session‑Idle‑Timeout response header uses minute unit, fallback to 30min if missing
			idleHeaderVal := response.Header.Get(consts.HeaderXSessionIdleTimeout)
			idleTimeoutMs := parseXSessionIdleTimeout(idleHeaderVal, sid, defaultIdleMin, klog)

			sess := redisx.AgwSession{
				IsLogin:             true,
				XSubjectToken:       XSubjectToken,
				RefreshPointTs:      refreshPointTs,
				TokenExpireTs:       tokenExpireAtMs,
				Refreshing:          0,
				IdleTimeoutMs:       idleTimeoutMs, // Independent idle timeout for each session
				LastAccessTs:        nowMs,         // Use login timestamp as initial access time
				LastCookieRefreshTs: nowMs,         // Initialize cookie refresh timestamp at login
			}

			err = redisx.Manager.AgwSession.Set(ctx, sid, sess, sttl)
			if err != nil {
				klog.Errorf("session save to redis err: %v sid=%s", err, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}

			session.Set(consts.SessionKeyIsLogin, true)
			session.Set(consts.RedisSessionKey, sid)
			if err = session.Save(); err != nil {
				klog.Errorf("session save err: %v sid=%s", err, sid)
				c.JSON(http.StatusInternalServerError, answer.NewResMessage(answer.EcodeSaveSessionError, "Internal Server Error", nil))
				return
			}
			klog.Infof("log in success. sid=[%s], idleTimeoutMs=%d", sid, idleTimeoutMs)
		}

		c.Status(response.StatusCode())
		response.Header.VisitAll(func(key, value []byte) {
			k := string(key)
			if blockRespHeaders[k] {
				return
			}
			c.Response.Header.Set(k, string(value))
		})

		c.Response.SetBody(response.Body())
	}
}
