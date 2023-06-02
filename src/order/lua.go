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

const luaPrepareCkTx = `
-- k1: user_id; k2: paid; k3: cart; k4: tx_id; k5: tx_state
-- a1: tx_id; a2: TxPreparing
-- return:
--          false
--          {0, user_id, paid, {}}
--          {1, user_id, paid, {item_id, amount, item_id, amount, ...}}

local user_id = redis.call('GET', KEYS[1])
if user_id == nil then
    return false
end

local paid = redis.call('GET', KEYS[2])
if paid == nil then
    return false
end

-- if already paid, the paying tx will be locked by the transaction id
-- aborted transaction will delete this key
local locked = redis.call('SET', KEYS[4], ARGV[1], 'NX')
if locked == nil then
    -- lock failed
    return {0, user_id, paid, {}}
end

redis.call('SET', KEYS[5], ARGV[2])

local cart = redis.call('HGETALL', KEYS[3])

return {1, user_id, paid, cart}
`

const luaAcknowledgeCkTx = `
-- k1: tx_id; k2: tx_state
-- a1: tx_id; a2: TxAcknowledged

local locked = redis.call('GET', KEYS[1])
if locked ~= ARGV[1] then
    -- not locked by this tx
    return {err = "error tx_id"}
end
redis.call('SET', KEYS[2], ARGV[2])
return true
`

const luaCommitCkTx = `
-- k1: tx_id; k2: tx_state; k3: paid
-- a1: tx_id; a2: TxCommitted

local locked = redis.call('GET', KEYS[1])
if locked ~= ARGV[1] then
    -- not locked by this tx
    return {err = "error tx_id"}
end
local paid = redis.call('SET', KEYS[3], 1, 'XX')
if paid == nil then
    -- no such order
    return {err = "error order_id"}
end
redis.call('SET', KEYS[2], ARGV[2])
return true
`

const luaAbortCkTx = `
-- k1: tx_id; k2: tx_state
-- a1: tx_id; a2: TxAborted

local locked = redis.call('GET', KEYS[1])
if locked ~= ARGV[1] then
    -- not locked by this tx
    return {err = "error tx_id"}
end
redis.call('DEL', KEYS[1])
redis.call('SET', KEYS[2], ARGV[2])
return true
`
