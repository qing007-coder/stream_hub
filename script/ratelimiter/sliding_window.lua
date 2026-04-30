-- KEYS[1]: 资源标识 (如 user_id 或 api_path)
-- ARGV[1]: 当前时间戳 (毫秒)
-- ARGV[2]: 窗口大小 (毫秒，例如 1000 代表 1 秒)
-- ARGV[3]: 窗口内最大允许请求数 (限流阈值)
-- ARGV[4]: 唯一成员标识 (防止同一毫秒的请求被覆盖)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_size = tonumber(ARGV[2])
local max_count = tonumber(ARGV[3])
local member_id = ARGV[4]

-- 计算窗口起点并清理过期数据
local window_start = now - window_size
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

-- 统计当前窗口内的请求总数
local current_count = redis.call('ZCARD', key)

-- 决策：是否允许本次请求
if current_count < max_count then
    -- 允许：将本次请求存入 ZSET
    redis.call('ZADD', key, now, member_id)
    -- 设置过期时间，防止冷 key 堆积 (略大于窗口大小)
    redis.call('EXPIRE', key, math.ceil(window_size / 1000) + 2)
    return true -- 代表准许
else
    -- 拦截：不写入数据
    return false -- 代表限流
end