package mw

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/sessions/redis"

	"apigw/src/config"
)

// 初始化Session
type redisCfg struct {
	addr          string
	passwd        string
	db            string
	sessionMaxAge int
	sKey          string
}

// 初始化redis配置
// redis DB default "0"
func NewRedisCfg(addr, passwd, db string, sma int) *redisCfg {
	if db == "" {
		log.Println(`Use the default db one ("0")`)
		db = "0"
	}
	return &redisCfg{
		addr:          addr,
		passwd:        passwd,
		db:            db,
		sessionMaxAge: sma,
		sKey:          "wderqeyJ2Y29kZSI6ImxkbG4ifQ.Xr-Lbg.ojkAcx7BZx7590luvEIvhYASA_8",
	}
}

// 使用redis保存session
func CreateStoreRedis(r *config.Redis, sessionMaxAge int) (app.HandlerFunc, error) {
	rcfg := NewRedisCfg(r.Addr, r.Password, r.Db, sessionMaxAge)
	// 连接redis
	store, err := redis.NewStoreWithDB(10, "tcp", rcfg.addr, rcfg.passwd, rcfg.db, []byte(rcfg.sKey))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{
		MaxAge: rcfg.sessionMaxAge,
		Path:   "/",
	})
	return sessions.New("session", store), nil
}

// 使用Cookie保存session
func CreateStoreCookie(sessionMaxAge int) (app.HandlerFunc, error) {
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		MaxAge: sessionMaxAge,
		Path:   "/",
	})
	return sessions.New("session", store), nil
}

// 初始化Session
func InitSession(h *server.Hertz, r *config.Redis) {
	// session设置超时时间(秒), 30min
	var sessionMaxAge int = 60 * 30
	var session app.HandlerFunc

	// 如果没有配置redis，则使用Cookie
	if r.Addr != "" {
		session, _ = CreateStoreRedis(r, sessionMaxAge)
		h.Use(session)
	} else {
		log.Println("redis is not configured. Use cookies to save drawing sessions.")
		session, _ = CreateStoreCookie(sessionMaxAge)
		h.Use(session)
	}
}
