package session

import (
	"apigw/src/cfgtypes"
	"apigw/src/pkg/consts"
	"apigw/src/slog"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/sessions/redis"
)

const (
	CookieName           = "session"
	DefaultSessionMaxAge = consts.CookieMaxAgeSec
	RedisNetwork         = "tcp"
	RedisSecretKey       = "wderqeyJ2Y29kZSI6ImxkbG4ifQ.Xr-Lbg.ojkAcx7BZx7590luvEIvhYASA_8"
	CookieSecret         = "secret"
	RedisPoolSize        = 10
)

var GlobalSessionOption = sessions.Options{
	MaxAge:   DefaultSessionMaxAge,
	Path:     "/",
	HttpOnly: true,
	SameSite: http.SameSiteStrictMode,
}

type redisCfg struct {
	addr          string
	network       string
	passwd        string
	db            string
	sessionMaxAge int
	sKey          string
}

// newRedisCfg 初始化redis配置，db默认0
func newRedisCfg(addr, passwd string, db, sma int) *redisCfg {
	return &redisCfg{
		addr:          addr,
		passwd:        passwd,
		db:            strconv.Itoa(db),
		sessionMaxAge: sma,
		network:       RedisNetwork,
		sKey:          RedisSecretKey,
	}
}

func CreateStoreRedis(r *cfgtypes.Redis, sessionMaxAge int) (app.HandlerFunc, error) {
	klog := slog.GetGlobal()
	rcfg := newRedisCfg(r.Addr, r.Password, r.Db, sessionMaxAge)
	klog.Infof("session redis connect: %s://%s/%s", rcfg.network, rcfg.addr, rcfg.db)

	store, err := redis.NewStoreWithDB(RedisPoolSize, rcfg.network, rcfg.addr, rcfg.passwd, rcfg.db, []byte(rcfg.sKey))
	if err != nil {
		klog.Fatalf("redis session create failed: %v", err)
		panic(err)
	}

	opt := GlobalSessionOption
	opt.MaxAge = rcfg.sessionMaxAge
	store.Options(opt)

	return sessions.New(CookieName, store), nil
}

// CreateStoreCookie 客户端Cookie存储Session
func CreateStoreCookie(sessionMaxAge int) (app.HandlerFunc, error) {
	klog := slog.GetGlobal()
	klog.Info("redis is not configured, use cookie store for session")

	store := cookie.NewStore([]byte(CookieSecret))
	opt := GlobalSessionOption
	opt.MaxAge = sessionMaxAge
	store.Options(opt)

	return sessions.New(CookieName, store), nil
}

// InitSession 全局会话初始化入口
func InitSession(h *server.Hertz, r *cfgtypes.Redis) {
	var sessionHandler app.HandlerFunc
	if r.Addr != "" {
		sessionHandler, _ = CreateStoreRedis(r, DefaultSessionMaxAge)
	} else {
		sessionHandler, _ = CreateStoreCookie(DefaultSessionMaxAge)
	}
	h.Use(sessionHandler)
}
