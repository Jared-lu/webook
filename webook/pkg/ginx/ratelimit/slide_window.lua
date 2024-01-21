-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber( ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口的起始时间，是从当前时间往前回溯窗口大小的时间
local min = now - window
-- 不在当前窗口的删掉
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
-- 当前窗口有多少个
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
-- local cnt = redis.call('ZCOUNT', key, min, '+inf')
if cnt >= threshold then
    -- 当前的请求数量已经大于窗口的大小，执行限流
    return "true"
else
    -- 把当前请求加到key上
    -- 把 score 和 value 都设置成 now
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return "false"
end