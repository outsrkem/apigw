package userauth

import (
	"apigw/src/pkg/answer"
	"apigw/src/pkg/consts"
	"apigw/src/pkg/redisx"
	"apigw/src/slog"
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
)

// UiasLogout 登出接口
func UiasLogout() func(context.Context, *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		session := sessions.Default(c)

		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get(consts.SessionKeyIsLogin)))
		if !isLogin {
			klog.Warnf("logout failed: current user not logged in")
			c.JSON(400, answer.NewResMessage(answer.EcodeOkay, "user not logged in.", nil))
			return
		}

		sid, ok := session.Get(consts.RedisSessionKey).(string)
		if !ok {
			klog.Warnf("sid not found in session")
			c.JSON(401, answer.ResBody(answer.EcodeNotLogIn, "user not logged in.", nil))
			return
		}

		expireOpt := sessions.Options{MaxAge: 0, Path: "/", HttpOnly: true}
		session.Options(expireOpt)
		session.Clear()
		if err := session.Save(); err != nil {
			klog.Errorf("logout session save error: %v", err)
			c.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "Logout failed", nil))
			return
		}

		klog.Debugf("logout try delete redis session, [%s]", sid)
		err := redisx.Manager.AgwSession.Delete(context.Background(), sid)
		if err != nil {
			klog.Errorf("failed to delete session info from Redis, [%s] err: %v", sid, err)
		} else {
			klog.Infof("logout redis session deleted ok, [%s]", sid)
		}

		klog.Infof("user logout success, cookie expired, [%s]", sid)
		c.JSON(200, answer.NewResMessage(answer.EcodeOkay, "Logout successful", nil))
	}
}
