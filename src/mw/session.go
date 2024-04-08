package mw

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/sessions/redis"

	"apigw/src/config"
)

// 初始化Session
// 如果配置了redis，则使用redis，没配置则使用cookie
func InitSession(h *server.Hertz, r *config.Redis) {

	var redisAddr string = r.Host + ":" + r.Port
	var redisPasswd string = r.Password
	var redisKey string = "wderqeyJ2Y29kZSI6ImxkbG4ifQ.Xr-Lbg.ojkAcx7BZx7590luvEIvhYASA_8"
	var sessionMaxAge int = 60 * 30 // session设置超时时间(秒) ,30min
	if r.Host != "" {
		log.Println("Save drawing sessions using redis.")
		log.Println("redis => ", redisAddr)

		// 连接redis
		store, err := redis.NewStore(10, "tcp", redisAddr, redisPasswd, []byte(redisKey))
		if err != nil {
			panic(err)
		}

		store.Options(sessions.Options{
			MaxAge: sessionMaxAge,
			Path:   "/",
		})
		h.Use(sessions.New("session", store))
	} else {
		log.Println("redis is not configured. Use cookies to save drawing sessions.")
		store := cookie.NewStore([]byte("secret"))
		store.Options(sessions.Options{
			MaxAge: sessionMaxAge, // 30min
			Path:   "/",
		})
		h.Use(sessions.New("session", store))
	}

}
