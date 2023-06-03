package stock

const luaPrepareCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxPreparing
-- return: state

local state = redis.call('GET', KEYS[1])
if state ~= nil then
    return state
end
-- new tx
redis.call('DEL', KEYS[2])  -- delete locked data
redis.call('SET', KEYS[1], ARGV[1])
return ''
`

const luaPrepareCkTxMove = `
-- k1: tx_state; k2: stock; k3: price; k4: tx_lk
-- a1: TxPreparing; a2: amount; a3: item_itemId; a4: 'price'
-- return: state / nil / error

local amount = tonumber(ARGV[2])
if amount == nil then
    return {err = "amount is nan"}
end

local stock = redis.call('GET', KEYS[2])
if stock == nil then
    return false
end
stock = tonumber(stock)
if not (stock ~= nil and stock - amount >= 0) then
    return false
end

local price = redis.call('GET', KEYS[3])
if price == nil then
    return false
end
price = tonumber(price)
if price == nil then
    return false
end


local state = redis.call('GET', KEYS[1])
if state == nil then
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxPreparing
    return state
end

redis.call('SET', KEYS[2], stock - amount)
redis.call('HINCRBY', KEYS[4], ARGV[3], amount)
redis.call('HINCRBY', KEYS[4], ARGV[4], price * amount)
return state
`

// todo
const luaAcknowledgeCkTx = `
`

const luaCommitCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxAcknowledged; a2: TxCommitted; a3: 'price'
-- return: state / nil

local state = redis.call('GET', KEYS[1])
if state == nil then
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxAcknowledged
    return state
end

-- release space
local price = redis.call('HGET', KEYS[2], ARGV[3])
redis.call('DEL', KEYS[2])
redis.call('HSET', KEYS[2], ARGV[3], price)

redis.call('SET', KEYS[1], ARGV[2])
return state
`
