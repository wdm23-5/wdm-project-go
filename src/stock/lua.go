package stock

const luaPrepareCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxPreparing; a2: 'price'
-- return: state

local state = redis.call('GET', KEYS[1])
if state ~= false then
    return state
end

-- new tx

-- reset locked data
redis.call('DEL', KEYS[2])
redis.call('HSET', KEYS[2], ARGV[2], 0)

redis.call('SET', KEYS[1], ARGV[1])
return ''
`

// incremental
const luaPrepareCkTxMove = `
-- k1: tx_state; k2: stock; k3: price; k4: tx_lk
-- a1: TxPreparing; a2: amount; a3: item_itemId; a4: 'price'
-- return: state / nil / error

local amount = tonumber(ARGV[2])
if amount == nil then
    return {err = "amount is nan"}
end

local state = redis.call('GET', KEYS[1])
if state == false then
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxPreparing
    return state
end

local stock = redis.call('GET', KEYS[2])
stock = tonumber(stock)
if not (stock ~= nil and stock - amount >= 0) then
    return false
end

local price = redis.call('GET', KEYS[3])
price = tonumber(price)
if price == nil then
    return false
end

redis.call('SET', KEYS[2], stock - amount)
redis.call('HSET', KEYS[4], ARGV[3], amount)
redis.call('HINCRBY', KEYS[4], ARGV[4], price * amount)
return state
`

const luaAcknowledgeCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxPreparing; a2: TxAcknowledged; a3: 'price'
-- return: {state, price} / err

local state = redis.call('GET', KEYS[1])
if state == false then
    return {'', ''}
end
if state ~= ARGV[1] then
    -- tx not TxPreparing
    return {state, ''}
end

local price = redis.call('HGET', KEYS[2], ARGV[3])
redis.call('SET', KEYS[1], ARGV[2])
return {state, price}
`

const luaCommitCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxAcknowledged; a2: TxCommitted; a3: 'price'
-- return: state

local state = redis.call('GET', KEYS[1])
if state == false then
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

const luaAbortCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxPreparing; a2: TxAcknowledged; a3: TxAborted; a4: 'price'
-- return: state

local state = redis.call('GET', KEYS[1])
if state == false then
    -- fast abort
    redis.call('SET', KEYS[1], ARGV[3])
    return ''
end
if not (state == ARGV[1] or state == ARGV[2]) then
    -- tx not TxPreparing or TxAcknowledged or nil
    return state
end

redis.call('HDEL', KEYS[2], ARGV[4])

redis.call('SET', KEYS[1], ARGV[3])
return state
`

// incremental
const luaAbortCkTxRollback = `
-- k1: tx_state; k2: tx_lk; k3: stock
-- a1: TxAborted; a2: item_itemId
-- return: state / nil

local state = redis.call('GET', KEYS[1])
if state == false then
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxAborted
    return state
end

local amount = redis.call('HGET', KEYS[2], ARGV[2])
amount = tonumber(amount)
if amount == nil then
    redis.call('HDEL', KEYS[2], ARGV[2])
    return false
end

redis.call('INCRBY', KEYS[3], amount)
redis.call('HDEL', KEYS[2], ARGV[2])
return state
`
