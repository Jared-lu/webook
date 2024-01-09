local key=KEYS[1]
-- 用户输入的验证码
local expectedCode =ARGV[1]
-- 系统生成的验证码
local code = redis.call("get",key)
-- 验证码剩余可校验次数
local cntKey=key..":cnt"
local cnt=tonumber(redis.call("get",cntKey))
-- 验证码的验证次数用完了
if cnt==nil or cnt<=0 then
    return -1
    -- 验证码正确
elseif expectedCode ==code then
    -- 将该验证码置为无效
    redis.call("set",cntKey,-1)
    return 0
else
    -- 输错了
    -- 可验证次数减1
    redis.call("decr",cntKey,-1)
    return -2
    
end