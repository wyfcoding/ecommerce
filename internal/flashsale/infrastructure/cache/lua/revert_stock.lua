-- revert_stock.lua
-- Keys: {flashsale:id}
-- Args: user_id, quantity

local key_tag = KEYS[1]
local user_id = ARGV[1]
local quantity = tonumber(ARGV[2])

local stock_key = key_tag .. ":stock"
local bought_key = key_tag .. ":bought"

-- 1. Revert stock
redis.call("incrby", stock_key, quantity)

-- 2. Revert bought count
local bought = tonumber(redis.call("hget", bought_key, user_id) or "0")
if bought >= quantity then
    redis.call("hincrby", bought_key, user_id, -quantity)
else
    redis.call("hdel", bought_key, user_id)
end

return 1
