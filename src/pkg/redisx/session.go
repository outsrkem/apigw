package redisx

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

//nolint:revive:exported
const (
	KeyAgwSessionData        = "agw:session:data:%s"
	FieldIsLogin             = "is_login"
	FieldXSubjectToken       = "x_subject_token"
	FieldRefreshPointTs      = "refresh_point_ts"
	FieldTokenExpireTs       = "token_expire_ts"
	FieldRefreshing          = "refreshing"
	FieldIdleTimeoutMs       = "idle_timeout_ms"        // Unit: millisecond
	FieldLastAccessTs        = "last_access_ts"         // Unit: millisecond
	FieldLastCookieRefreshTs = "last_cookie_refresh_ts" // Cookie refresh timestamp, unit: millisecond
	RefreshingLockTimeoutMs  = 30 * 1000                // Refresh lock timeout in milliseconds
)

var (
	scriptUpdateSingleFieldNoExpire   = redis.NewScript(luaUpdateSingleFieldNoExpire)
	scriptUpdateSingleFieldWithExpire = redis.NewScript(luaUpdateSingleFieldWithExpire)
	scriptUpdateTokenFields           = redis.NewScript(luaUpdateTokenFields)
	scriptTryAcquireRefreshingLock    = redis.NewScript(luaTryAcquireRefreshingLock)
)

// AgwSession redis hash mapping struct
type AgwSession struct {
	IsLogin             bool
	XSubjectToken       string
	RefreshPointTs      int64 // Pre‑refresh trigger timestamp in milliseconds
	TokenExpireTs       int64 // Token expire timestamp in milliseconds
	Refreshing          int64 // 0=no lock; non‑zero means lock start timestamp in milliseconds
	IdleTimeoutMs       int64 // Session idle timeout in milliseconds
	LastAccessTs        int64 // User last access timestamp in milliseconds
	LastCookieRefreshTs int64 // Last cookie refresh timestamp in milliseconds
}

// BuildAgwSessionKey build redis session key
func BuildAgwSessionKey(sid string) string {
	return fmt.Sprintf(KeyAgwSessionData, sid)
}

// AgwSessionRepo redis repository for agw session
type AgwSessionRepo struct {
	client *redis.Client
}

func NewAgwSessionRepo(client *redis.Client) *AgwSessionRepo {
	return &AgwSessionRepo{client: client}
}

// Set create or update session and set key ttl
func (r *AgwSessionRepo) Set(ctx context.Context, sid string, sess AgwSession, ttl time.Duration) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	data := map[string]any{
		FieldIsLogin:             strconv.FormatBool(sess.IsLogin),
		FieldXSubjectToken:       sess.XSubjectToken,
		FieldRefreshPointTs:      strconv.FormatInt(sess.RefreshPointTs, 10),
		FieldTokenExpireTs:       strconv.FormatInt(sess.TokenExpireTs, 10),
		FieldRefreshing:          strconv.FormatInt(sess.Refreshing, 10),
		FieldIdleTimeoutMs:       strconv.FormatInt(sess.IdleTimeoutMs, 10),
		FieldLastAccessTs:        strconv.FormatInt(sess.LastAccessTs, 10),
		FieldLastCookieRefreshTs: strconv.FormatInt(sess.LastCookieRefreshTs, 10),
	}
	pipe := r.client.Pipeline()
	pipe.HMSet(ctx, key, data)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// Get load session by sid; returns redis.Nil if key does not exist
func (r *AgwSessionRepo) Get(ctx context.Context, sid string) (*AgwSession, error) {
	if r == nil {
		return nil, fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, redis.Nil
	}
	sess := &AgwSession{}
	sess.IsLogin, _ = strconv.ParseBool(res[FieldIsLogin])
	sess.XSubjectToken = res[FieldXSubjectToken]
	sess.RefreshPointTs, _ = strconv.ParseInt(res[FieldRefreshPointTs], 10, 64)
	sess.TokenExpireTs, _ = strconv.ParseInt(res[FieldTokenExpireTs], 10, 64)
	sess.Refreshing, _ = strconv.ParseInt(res[FieldRefreshing], 10, 64)
	sess.IdleTimeoutMs, _ = strconv.ParseInt(res[FieldIdleTimeoutMs], 10, 64)
	sess.LastAccessTs, _ = strconv.ParseInt(res[FieldLastAccessTs], 10, 64)
	sess.LastCookieRefreshTs, _ = strconv.ParseInt(res[FieldLastCookieRefreshTs], 10, 64)
	return sess, nil
}

// Delete remove session on logout
func (r *AgwSessionRepo) Delete(ctx context.Context, sid string) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	return r.client.Del(ctx, key).Err()
}

// SetRefreshingLock set refresh lock with current millisecond timestamp
func (r *AgwSessionRepo) SetRefreshingLock(ctx context.Context, sid string, nowMs int64) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	return r.client.HSet(ctx, key, FieldRefreshing, strconv.FormatInt(nowMs, 10)).Err()
}

// ClearRefreshingLock clear refresh lock and set value to 0
func (r *AgwSessionRepo) ClearRefreshingLock(ctx context.Context, sid string) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	return r.client.HSet(ctx, key, FieldRefreshing, "0").Err()
}

// UpdateToken used by async refresh goroutine: update token, refresh point and token expire timestamp; renew redis key ttl
// ttl: new key expiration duration (remaining token lifetime plus safety margin)
func (r *AgwSessionRepo) UpdateToken(ctx context.Context, sid, newToken string, newRefreshPointTs, newTokenExpireTs int64, ttl time.Duration) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	ttlSec := int64(ttl.Seconds())
	_, err := scriptUpdateTokenFields.Run(ctx, r.client, []string{key},
		FieldXSubjectToken, newToken,
		FieldRefreshPointTs, strconv.FormatInt(newRefreshPointTs, 10),
		FieldTokenExpireTs, strconv.FormatInt(newTokenExpireTs, 10),
		ttlSec,
	).Result()
	return err
}

// UpdateLastAccess update session last‑access timestamp
func (r *AgwSessionRepo) UpdateLastAccess(ctx context.Context, sid string, nowMs int64) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	_, err := scriptUpdateSingleFieldNoExpire.Run(ctx, r.client, []string{key},
		FieldLastAccessTs,
		strconv.FormatInt(nowMs, 10),
	).Result()
	return err
}

// UpdateLastCookieRefreshTs update cookie refresh timestamp in milliseconds
func (r *AgwSessionRepo) UpdateLastCookieRefreshTs(ctx context.Context, sid string, tsMs int64) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	_, err := scriptUpdateSingleFieldNoExpire.Run(ctx, r.client, []string{key},
		FieldLastCookieRefreshTs,
		strconv.FormatInt(tsMs, 10),
	).Result()
	return err
}

// UpdateLastAccessAndExpire update last‑access timestamp and slide redis key expiration time
func (r *AgwSessionRepo) UpdateLastAccessAndExpire(ctx context.Context, sid string, nowMs int64, ttl time.Duration) error {
	if r == nil {
		return fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	ttlSec := int64(ttl.Seconds())
	_, err := scriptUpdateSingleFieldWithExpire.Run(ctx, r.client, []string{key},
		FieldLastAccessTs,
		strconv.FormatInt(nowMs, 10),
		ttlSec,
	).Result()
	return err
}

// TryAcquireRefreshingLock atomically try to acquire token refresh lock
// return true=lock acquired; false=lock held or session not exists
func (r *AgwSessionRepo) TryAcquireRefreshingLock(ctx context.Context, sid string, nowMs, lockTimeoutMs int64) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("AgwSessionRepo is nil")
	}
	key := BuildAgwSessionKey(sid)
	res, err := scriptTryAcquireRefreshingLock.Run(ctx, r.client, []string{key},
		strconv.FormatInt(nowMs, 10),
		strconv.FormatInt(lockTimeoutMs, 10),
	).Result()
	if err != nil {
		return false, err
	}
	val, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("script return invalid type")
	}
	return val == 1, nil
}
