-- 该代码的返回值是我们自己定义的，不是redis ttl命令的返回值

-- 拿到外部传入的Key，phone_code:login:152xxxxxxxx
local key=KEYS[1]
-- 拼接字符
-- 验证次数，一个验证码最多重复三次，这个记录了验证了几次，对应的还可以验证几次
-- phone_code:login:152xxxxxxxx:cnt
local cntKey=key..":cnt"
-- 你的验证码
local val=ARGV[1]
-- 过期时间
-- 调用ttl命令
local ttl=tonumber(redis.call("ttl",key))
-- -1 Key存在，但没有过期时间，说明系统异常
if ttl == -1 then
    return -2
    -- -2 key不存在，小于540说明验证码已经过去了一分钟
elseif ttl==-2 or ttl<540 then
    -- 存储验证码
    redis.call("set",key,val)
    -- 设置过期时间
    redis.call("expire",key,600)
    -- 设置验证次数，最多允许3次
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
    return 0
else
    -- 发送太频繁
    return -1
end

