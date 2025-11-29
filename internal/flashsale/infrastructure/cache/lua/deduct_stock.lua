-- deduct_stock.lua
-- Keys: {flashsale:id}
-- Args: user_id, quantity, limit_per_user

local key_tag = KEYS[1]
local user_id = ARGV[1]
local quantity = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local stock_key = key_tag .. ":stock"
local bought_key = key_tag .. ":bought"

-- 1. Check stock
local stock = tonumber(redis.call("get", stock_key) or "0")
if stock < quantity then
    return -1 -- Not enough stock
end

-- 2. Check user limit
local bought = tonumber(redis.call("hget", bought_key, user_id) or "0")
if (bought + quantity) > limit then
    return -2 -- Exceed limit
end

-- 3. Deduct stock and update bought count
redis.call("decrby", stock_key, quantity)
redis.call("hincrby", bought_key, user_id, quantity)

return 1 -- Success
