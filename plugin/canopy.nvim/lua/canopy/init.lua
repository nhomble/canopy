local cache = require("canopy.cache")
local client = require("canopy.client")
local server_mod = require("canopy.server")

local M = {}

M.config = {
  base_url = nil, -- nil = auto-detect per repo; set to override for all repos
  binary = "canopy", -- path to canopy binary (if not on PATH, use absolute path)
}

local setup_done = false
local last_cursor_file = nil

--- Merge user options and register commands + autocmds.
--- @param opts table|nil
function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})

  if setup_done then
    -- Config updated, but commands/autocmds already registered
    if M.config.binary then
      server_mod.set_binary(M.config.binary)
    end
    return
  end
  setup_done = true
  M._setup_done = true

  -- Configure binary path
  if M.config.binary then
    server_mod.set_binary(M.config.binary)
  end

  -- Commands
  vim.api.nvim_create_user_command("CanopyContext", function()
    local bufnr = vim.api.nvim_get_current_buf()
    local url = cache.base_url_for(bufnr, M.config.base_url)
    if not url then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    local ctx = cache.get(bufnr, url)
    if not ctx then
      vim.notify("canopy: no context for this file", vim.log.levels.WARN)
      return
    end
    require("canopy.ui").show_context(ctx)
  end, { desc = "Show architectural context for current file" })

  vim.api.nvim_create_user_command("CanopyFlow", function()
    local bufnr = vim.api.nvim_get_current_buf()
    local url = cache.base_url_for(bufnr, M.config.base_url)
    if not url then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    require("canopy.flows").for_current_file(url)
  end, { desc = "Show flows through current file" })

  vim.api.nvim_create_user_command("CanopyOpen", function()
    local bufnr = vim.api.nvim_get_current_buf()
    local url = cache.base_url_for(bufnr, M.config.base_url)
    if not url then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    local cmd
    if vim.fn.has("mac") == 1 then
      cmd = { "open", url }
    elseif vim.fn.has("wsl") == 1 then
      cmd = { "wslview", url }
    else
      cmd = { "xdg-open", url }
    end
    vim.fn.jobstart(cmd, { detach = true })
  end, { desc = "Open canopy web UI in browser" })

  vim.api.nvim_create_user_command("CanopyStart", function()
    local root = cache.find_project_root(vim.api.nvim_get_current_buf())
    if not root then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    server_mod.start(root)
  end, { desc = "Start canopy server for current project" })

  vim.api.nvim_create_user_command("CanopyStop", function()
    local root = cache.find_project_root(vim.api.nvim_get_current_buf())
    if not root then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    server_mod.stop(root)
  end, { desc = "Stop canopy server for current project" })

  vim.api.nvim_create_user_command("CanopyRestart", function()
    local root = cache.find_project_root(vim.api.nvim_get_current_buf())
    if not root then
      vim.notify("canopy: not in a canopy project", vim.log.levels.WARN)
      return
    end
    server_mod.restart(root)
  end, { desc = "Restart canopy server for current project" })

  -- Auto-start server and prefetch context on BufEnter
  vim.api.nvim_create_autocmd("BufEnter", {
    group = vim.api.nvim_create_augroup("canopy", { clear = true }),
    callback = function(ev)
      local root = cache.find_project_root(ev.buf)
      if not root then
        return
      end

      -- Ensure server is running (auto-start if needed, never block editor)
      pcall(server_mod.ensure, root)

      local url = cache.base_url_for(ev.buf, M.config.base_url)
      if not url then
        return
      end

      cache.invalidate(ev.buf)
      -- Prefetch silently (ignore errors)
      cache.get(ev.buf, url)

      -- Report cursor position to server for live web UI sync
      local rel = cache.relative_path(ev.buf)
      if rel and rel ~= last_cursor_file then
        last_cursor_file = rel
        client.put_async(url .. "/cursor?file=" .. vim.uri_encode(rel, "rfc2396"))
      end
    end,
  })
end

--- Return a statusline string for the current buffer.
--- Suitable for lualine: lualine_x = { require('canopy').statusline }
--- @return string
function M.statusline()
  local bufnr = vim.api.nvim_get_current_buf()
  local root = cache.find_project_root(bufnr)
  if not root then
    return ""
  end

  local status = server_mod.status(root)
  local port_str = status.running and tostring(status.port) or "off"

  local url = cache.base_url_for(bufnr, M.config.base_url)
  if not url then
    return "canopy:" .. port_str
  end
  local ctx = cache.get(bufnr, url)
  if not ctx then
    return "canopy:" .. port_str
  end

  local parts = {}
  if ctx.component and ctx.component.name then
    table.insert(parts, ctx.component.name)
  end
  if ctx.layer and ctx.layer ~= "" then
    table.insert(parts, ctx.layer)
  end
  if ctx.archetype and ctx.archetype.category then
    table.insert(parts, ctx.archetype.category)
  end

  if #parts == 0 then
    return "canopy:" .. port_str
  end

  return "canopy:" .. port_str .. " [" .. table.concat(parts, " | ") .. "]"
end

return M
