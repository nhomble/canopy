local port_mod = require("canopy.port")

local M = {}

-- Binary path, set via M.set_binary() from init.lua setup
local binary = "canopy"

-- Per-root server state: root_path → { job_id = N, port = N }
local servers = {}

--- Set the binary path for the canopy CLI.
--- @param path string
function M.set_binary(path)
  binary = path
end

--- Get the deterministic port for a repo root.
--- @param root string Absolute path (no trailing slash)
--- @return number
function M.port_for(root)
  return port_mod.for_root(root)
end

--- Check if the server for a given root is responding.
--- @param root string
--- @return boolean
function M.is_running(root)
  root = root:gsub("/$", "")
  local p = M.port_for(root)
  local output = vim.fn.system({
    "curl",
    "-s",
    "-o",
    "/dev/null",
    "-w",
    "%{http_code}",
    "--max-time",
    "1",
    "http://127.0.0.1:" .. p .. "/health",
  })
  return vim.v.shell_error == 0 and vim.trim(output) == "200"
end

--- Start the canopy server for a repo root.
--- @param root string Absolute path (no trailing slash)
--- @return boolean success
function M.start(root)
  root = root:gsub("/$", "")
  local p = M.port_for(root)

  -- Already running?
  if M.is_running(root) then
    return true
  end

  -- Check binary exists
  if vim.fn.executable(binary) ~= 1 then
    vim.notify("canopy: binary not found: " .. binary, vim.log.levels.WARN)
    return false
  end

  local job_id
  job_id = vim.fn.jobstart({
    binary,
    "serve",
    "--port",
    tostring(p),
  }, {
    cwd = root,
    on_exit = function()
      local state = servers[root]
      if state and state.job_id == job_id then
        servers[root] = nil
      end
    end,
  })

  if job_id <= 0 then
    vim.notify("canopy: failed to start server", vim.log.levels.ERROR)
    return false
  end

  servers[root] = { job_id = job_id, port = p }
  return true
end

--- Stop the canopy server for a repo root.
--- @param root string Absolute path (no trailing slash)
function M.stop(root)
  root = root:gsub("/$", "")
  local state = servers[root]
  if state then
    vim.fn.jobstop(state.job_id)
    servers[root] = nil
    vim.notify("canopy: server stopped", vim.log.levels.INFO)
    return
  end

  -- No tracked job, but server might be running externally — try kill via port
  local p = M.port_for(root)
  vim.fn.system({
    "curl",
    "-s",
    "--max-time",
    "1",
    "-X",
    "PUT",
    "http://127.0.0.1:" .. p .. "/shutdown",
  })
end

--- Restart the canopy server for a repo root.
--- @param root string
function M.restart(root)
  root = root:gsub("/$", "")
  M.stop(root)
  -- Small delay to let the port free up
  vim.defer_fn(function()
    M.start(root)
  end, 500)
end

--- Ensure the server is running, starting it if needed.
--- @param root string
--- @return boolean running
function M.ensure(root)
  root = root:gsub("/$", "")
  if M.is_running(root) then
    return true
  end
  return M.start(root)
end

--- Get status info for a repo root.
--- @param root string
--- @return table { running = bool, port = number }
function M.status(root)
  root = root:gsub("/$", "")
  return {
    running = M.is_running(root),
    port = M.port_for(root),
  }
end

return M
