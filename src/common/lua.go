package common

// atomically increase the value if the key already exists and the result will still be >= 0
const LuaIncrByIfGe0XX = `
local delta = tonumber(ARGV[1])
if delta == nil then
    return {err = "delta is nan"}
end
local value = redis.call('GET', KEYS[1])
value = tonumber(value)  -- tonumber(false) == nil
if value ~= nil and value + delta >= 0 then
    redis.call('SET', KEYS[1], value + delta)  -- assume OK
    return value + delta
end
return false
`
