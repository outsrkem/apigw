package redisx

// lua script:
// pre‑check key existence,
// return directly if session is deleted,
// prevent hash reconstruction from async updates
const (
	luaUpdateSingleFieldNoExpire = `
if redis.call('EXISTS', KEYS[1]) == 1 then
    redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
    return 1
end
return 0
`

	luaUpdateSingleFieldWithExpire = `
if redis.call('EXISTS', KEYS[1]) == 1 then
    redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
    redis.call('EXPIRE', KEYS[1], ARGV[3])
    return 1
end
return 0
`

	luaUpdateTokenFields = `
if redis.call('EXISTS', KEYS[1]) == 1 then
    redis.call('HMSET', KEYS[1], ARGV[1], ARGV[2], ARGV[3], ARGV[4], ARGV[5], ARGV[6])
    redis.call('EXPIRE', KEYS[1], ARGV[7])
    return 1
end
return 0
`

	luaTryAcquireRefreshingLock = `
if redis.call('EXISTS', KEYS[1]) == 1 then
    local refreshing = tonumber(redis.call('HGET', KEYS[1], 'refreshing') or "0")
    local now = tonumber(ARGV[1])
    local lockTimeout = tonumber(ARGV[2])
    if refreshing == 0 or (now - refreshing) > lockTimeout then
        redis.call('HSET', KEYS[1], 'refreshing', now)
        return 1
    end
end
return 0
`
)
