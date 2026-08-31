-- Atomic sliding-window rate limit check.
-- KEYS[1] = rate:{client_id}
-- ARGV[1] = now (ms, integer)
-- ARGV[2] = window (ms, integer)
-- ARGV[3] = limit (integer, max requests allowed within window)
-- ARGV[4] = req_id (unique member per request, e.g. uuid or now:counter)
--
-- Returns {allowed (0/1), remaining}

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local req_id = ARGV[4]

local clear_before = now - window
redis.call('ZREMRANGEBYSCORE', key, '-inf', clear_before)

local current_requests = redis.call('ZCARD', key)

if current_requests < limit then
    redis.call('ZADD', key, now, req_id)
    -- PEXPIRE in ms; window is already ms so this keeps the key from
    -- growing unbounded once a client goes idle.
    redis.call('PEXPIRE', key, window)
    return {1, limit - current_requests - 1}
else
    return {0, 0}
end
