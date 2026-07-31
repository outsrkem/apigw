package userauth

import (
	"apigw/src/pkg/answer"
	"apigw/src/slog"
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
)

// UiasLogout 登出接口
func UiasLogout() func(context.Context, *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromCtx(c)
		session := sessions.Default(ctx)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))

		if !isLogin {
			klog.Warn("logout failed: current user not logged in")
			ctx.JSON(400, answer.NewResMessage(answer.EcodeOkay, "User not logged in", nil))
			return
		}

		expireOpt := sessions.Options{
			MaxAge:   0,
			Path:     "/",
			HttpOnly: true,
		}

		session.Options(expireOpt)

		session.Clear()

		if err := session.Save(); err != nil {
			klog.Errorf("logout session save error: %v", err)
			ctx.JSON(500, answer.NewResMessage(answer.EcodeBackEndServiceError, "Logout failed", nil))
			return
		}

		klog.Info("user logout success, cookie expired")
		ctx.JSON(200, answer.NewResMessage(answer.EcodeOkay, "Logout successful", nil))
	}
}
