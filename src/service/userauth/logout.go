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

// UiasLogout 登出
func UiasLogout() func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromContext(ctx)
		session := sessions.Default(ctx)
		isLogin, _ := strconv.ParseBool(fmt.Sprint(session.Get("isLogin")))

		if !isLogin {
			klog.Warnf("user not logged in.")
			ctx.JSON(400, answer.NewResMessage(answer.EcodeOkay, "User not logged in", nil))
		}

		// session.Options(sessions.Options{
		// 	MaxAge: 0,
		// })
		session.Clear()
		if err := session.Save(); err != nil {
			klog.Errorf("save session err %v", err)
		}

		klog.Info("logout success.")
		ctx.JSON(200, answer.NewResMessage(answer.EcodeOkay, nil, nil))
	}
}
