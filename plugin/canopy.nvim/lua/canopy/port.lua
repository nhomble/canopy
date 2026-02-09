local bit = require("bit")

local M = {}

local FNV_OFFSET = 2166136261
local FNV_PRIME = 16777619
local PORT_BASE = 10000
local PORT_RANGE = 50000

--- Multiply two 32-bit integers mod 2^32 using 16-bit halves.
--- This avoids Lua double precision loss for large products.
--- @param a number
--- @param b number
--- @return number 32-bit result (signed, as returned by bit.band)
local function mul32(a, b)
  -- Ensure unsigned 32-bit inputs
  if a < 0 then
    a = a + 4294967296
  end
  if b < 0 then
    b = b + 4294967296
  end
  local a_lo = a % 65536
  local a_hi = (a - a_lo) / 65536
  local b_lo = b % 65536
  local b_hi = (b - b_lo) / 65536
  -- (a_hi*65536 + a_lo) * (b_hi*65536 + b_lo) mod 2^32
  -- = a_lo*b_lo + (a_lo*b_hi + a_hi*b_lo)*65536   (mod 2^32, dropping a_hi*b_hi*2^32)
  local lo = a_lo * b_lo
  local mid = (a_lo * b_hi + a_hi * b_lo) % 65536
  return bit.band(lo + mid * 65536, 0xFFFFFFFF)
end

--- Compute a deterministic port from a repo root path.
--- Uses FNV-1a 32-bit, matching the Go implementation in internal/server/port.go.
---
--- Cross-language test vectors (must match Go):
---   "/home/user/my-project"              → 50556
---   "/Users/nicolas/dev/codebase-viz"    → 35385
---   "/tmp/hexagonal-ddd"                 → 25612
---   "/tmp/test"                          → 17752
---
--- @param root string Absolute path to repo root (no trailing slash)
--- @return number port Port in range [10000, 59999]
function M.for_root(root)
  local h = FNV_OFFSET
  for i = 1, #root do
    h = bit.bxor(h, root:byte(i))
    h = mul32(h, FNV_PRIME)
  end
  -- bit ops return signed 32-bit in LuaJIT; convert to unsigned
  if h < 0 then
    h = h + 4294967296
  end
  return PORT_BASE + (h % PORT_RANGE)
end

return M
