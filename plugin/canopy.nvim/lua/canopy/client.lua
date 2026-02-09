local M = {}

--- Perform a synchronous GET request to the canopy server.
--- @param url string Full URL to fetch
--- @return table|nil data Parsed JSON response
--- @return string|nil err Error message on failure
function M.get(url)
  local output = vim.fn.system({ "curl", "-s", "--max-time", "2", url })
  if vim.v.shell_error ~= 0 then
    return nil, "curl failed (exit " .. vim.v.shell_error .. "): " .. output
  end

  local ok, decoded = pcall(vim.json.decode, output)
  if not ok then
    return nil, "JSON decode failed: " .. tostring(decoded)
  end

  return decoded, nil
end

--- Fire-and-forget async PUT request (non-blocking).
--- @param url string Full URL to PUT
function M.put_async(url)
  vim.fn.jobstart({ "curl", "-s", "-X", "PUT", "--max-time", "1", url }, { detach = true })
end

return M
