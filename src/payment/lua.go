package payment

const luaPrpThenAckAbtCkTx = `
-- k1: tx_state; k2: tx_lk; k3: credit
-- a1: TxAcknowledged; a2: TxAborted; a3: amount; a4: user_id
-- return: state / nil / error

local amount = tonumber(ARGV[3])
if amount == nil then
    return {err = "amount is nan"}
end

local state = redis.call('GET', KEYS[1])
if state ~= false then
    return state
end

-- new tx

redis.call('DEL', KEYS[2])

local credit = redis.call('GET', KEYS[3])
credit = tonumber(credit)
if not (credit ~= nil and credit - amount >= 0) then
    redis.call('SET', KEYS[1], ARGV[2])
    return false
end

redis.call('HSET', KEYS[2], ARGV[4], amount)
redis.call('SET', KEYS[3], credit - amount)

redis.call('SET', KEYS[1], ARGV[1])
return ''
`

const luaCommitCkTx = `
-- k1: tx_state; k2: tx_lk
-- a1: TxAcknowledged; a2: TxCommitted
-- return: state

local state = redis.call('GET', KEYS[1])
if state == false then
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxAcknowledged
    return state
end

redis.call('DEL', KEYS[2])

redis.call('SET', KEYS[1], ARGV[2])
return state
`

const luaAbtThenRollbackCkTx = `
-- k1: tx_state; k2: tx_lk; k3: credit
-- a1: TxAcknowledged; a2: TxAborted; a3: user_id
-- return: state / nil

local state = redis.call('GET', KEYS[1])
if state == false then
    -- fast abort
    redis.call('SET', KEYS[1], ARGV[2])
    return ''
end
if state ~= ARGV[1] then
    -- tx not TxAcknowledged or nil
    return state
end

local amount = redis.call('HGET', KEYS[2], ARGV[3])
amount = tonumber(amount)
if amount == nil then
    redis.call('DEL', KEYS[2])

    redis.call('SET', KEYS[1], ARGV[2])
    return false
end

redis.call('INCRBY', KEYS[3], amount)
redis.call('DEL', KEYS[2])

redis.call('SET', KEYS[1], ARGV[2])
return state
`
