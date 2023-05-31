package order

const luaHDecrIfGe0XX = `
local value = redis.call('HGET', KEYS[1], ARGV[1])
if value == nil then
    return false
end
value = tonumber(value)
if value ~= nil and value - 1 >= 0 then
    redis.call('HSET', KEYS[1], ARGV[1], value - 1)
	return value - 1
end
return false
`
