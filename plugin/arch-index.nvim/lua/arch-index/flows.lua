local cache = require("arch-index.cache")
local client = require("arch-index.client")
local ui = require("arch-index.ui")

local M = {}

--- Find the project root for the current buffer (for resolving relative paths).
--- @return string|nil
local function project_root()
  local bufpath = vim.api.nvim_buf_get_name(0)
  if bufpath == "" then
    return nil
  end
  local dir = vim.fn.fnamemodify(bufpath, ":h")
  while dir and dir ~= "/" do
    if vim.fn.isdirectory(dir .. "/.arch") == 1 then
      return dir
    end
    local parent = vim.fn.fnamemodify(dir, ":h")
    if parent == dir then
      break
    end
    dir = parent
  end
  return nil
end

--- Show flows through the current file, then let user pick one to see steps.
--- @param base_url string
function M.for_current_file(base_url)
  local bufnr = vim.api.nvim_get_current_buf()
  local ctx = cache.get(bufnr, base_url)
  if not ctx then
    vim.notify("arch-index: no context for this file", vim.log.levels.WARN)
    return
  end

  -- Collect flow IDs to query: from archetype and component
  local through_ids = {}
  if ctx.archetype and ctx.archetype.id then
    table.insert(through_ids, ctx.archetype.id)
  end
  if ctx.component and ctx.component.id then
    table.insert(through_ids, ctx.component.id)
  end

  if #through_ids == 0 then
    vim.notify("arch-index: no flows through this file", vim.log.levels.INFO)
    return
  end

  -- Fetch flows for each ID and deduplicate
  local seen = {}
  local flows = {}
  for _, id in ipairs(through_ids) do
    local data, err = client.get(base_url .. "/flows?through=" .. vim.uri_encode(id, "rfc2396"))
    if data and data.flows then
      for _, flow in ipairs(data.flows) do
        if not seen[flow.id] then
          seen[flow.id] = true
          table.insert(flows, flow)
        end
      end
    end
  end

  if #flows == 0 then
    vim.notify("arch-index: no flows through this file", vim.log.levels.INFO)
    return
  end

  -- If only one flow, show steps directly
  if #flows == 1 then
    M.show_steps(flows[1], base_url)
    return
  end

  -- Multiple flows: let user pick
  ui.pick_flow(flows, function(flow)
    M.show_steps(flow, base_url)
  end)
end

--- Show the steps of a flow with file navigation.
--- @param flow table Flow object { id, name, steps }
--- @param base_url string
function M.show_steps(flow, base_url)
  local archetype_files = cache.get_archetype_files(base_url)
  local root = project_root()
  ui.show_flow_steps(flow, archetype_files, root)
end

return M
