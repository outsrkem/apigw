package consts

import "time"

const (
	RedisSessionKey   = "sid"
	SessionKeyIsLogin = "isLogin"
)

const (
	CookieMaxAgeSec         = 3 * 24 * 60 * 60 // Cookie maximum lifetime in seconds
	CookieRenewThresholdSec = 2 * 24 * 60 * 60 // Renew cookie when remaining lifetime is less than this value(seconds)
	AgwSessionTTLMargin     = 10 * time.Minute // Extra ttl margin for redis agw session key
)
