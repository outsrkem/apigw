package session

import (
	"apigw/src/cfgtypts"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/sessions/redis"
)

// redis配置
type redisCfg struct {
	addr          string
	network       string
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
		network:       "tcp",
		sKey:          "wderqeyJ2Y29kZSI6ImxkbG4ifQ.Xr-Lbg.ojkAcx7BZx7590luvEIvhYASA_8",
	}
}

// CreateStoreRedis 使用redis保存session
func CreateStoreRedis(r *cfgtypts.Redis, sessionMaxAge int) (app.HandlerFunc, error) {
	rcfg := NewRedisCfg(r.Addr, r.Password, r.Db, sessionMaxAge)
	log.Println("session redis:", rcfg.network+"://"+r.Addr+"/"+r.Db)
	// 连接redis
	store, err := redis.NewStoreWithDB(10, rcfg.network, rcfg.addr, rcfg.passwd, rcfg.db, []byte(rcfg.sKey))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{
		MaxAge: rcfg.sessionMaxAge,
		Path:   "/",
	})
	return sessions.New("session", store), nil
}

// CreateStoreCookie 使用Cookie保存session
func CreateStoreCookie(sessionMaxAge int) (app.HandlerFunc, error) {
	log.Println("redis is not configured. Use cookies to save drawing sessions.")
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		MaxAge: sessionMaxAge,
		Path:   "/",
	})
	return sessions.New("session", store), nil
}

// InitSession 初始化Session
func InitSession(h *server.Hertz, r *cfgtypts.Redis) {
	// session设置超时时间(秒), 30min
	var sessionMaxAge int = 60 * 30
	var session app.HandlerFunc

	// 如果没有配置redis，则使用Cookie
	if r.Addr != "" {
		session, _ = CreateStoreRedis(r, sessionMaxAge)
	} else {
		session, _ = CreateStoreCookie(sessionMaxAge)
	}
	h.Use(session)
}
