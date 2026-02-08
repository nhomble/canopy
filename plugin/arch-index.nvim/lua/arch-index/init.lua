local cache = require("arch-index.cache")
local client = require("arch-index.client")

local M = {}

M.config = {
  base_url = "http://127.0.0.1:3451",
}

local setup_done = false
local last_cursor_file = nil

--- Merge user options and register commands + autocmds.
--- @param opts table|nil
function M.setup(opts)
  if setup_done then
    return
  end
  setup_done = true

  M.config = vim.tbl_deep_extend("force", M.config, opts or {})

  -- Commands
  vim.api.nvim_create_user_command("ArchContext", function()
    local ctx = cache.get(vim.api.nvim_get_current_buf(), M.config.base_url)
    if not ctx then
      vim.notify("arch-index: no context for this file", vim.log.levels.WARN)
      return
    end
    require("arch-index.ui").show_context(ctx)
  end, { desc = "Show architectural context for current file" })

  vim.api.nvim_create_user_command("ArchFlow", function()
    require("arch-index.flows").for_current_file(M.config.base_url)
  end, { desc = "Show flows through current file" })

  vim.api.nvim_create_user_command("ArchOpen", function()
    local url = M.config.base_url
    local cmd
    if vim.fn.has("mac") == 1 then
      cmd = { "open", url }
    elseif vim.fn.has("wsl") == 1 then
      cmd = { "wslview", url }
    else
      cmd = { "xdg-open", url }
    end
    vim.fn.jobstart(cmd, { detach = true })
  end, { desc = "Open arch-index web UI in browser" })

  -- Prefetch context on BufEnter
  vim.api.nvim_create_autocmd("BufEnter", {
    group = vim.api.nvim_create_augroup("arch-index", { clear = true }),
    callback = function(ev)
      cache.invalidate(ev.buf)
      -- Prefetch silently (ignore errors)
      cache.get(ev.buf, M.config.base_url)

      -- Report cursor position to server for live web UI sync
      local rel = cache.relative_path(ev.buf)
      if rel and rel ~= last_cursor_file then
        last_cursor_file = rel
        client.put_async(M.config.base_url .. "/cursor?file=" .. vim.uri_encode(rel, "rfc2396"))
      end
    end,
  })
end

--- Return a statusline string for the current buffer.
--- Suitable for lualine: lualine_x = { require('arch-index').statusline }
--- @return string
function M.statusline()
  local bufnr = vim.api.nvim_get_current_buf()
  local ctx = cache.get(bufnr, M.config.base_url)
  if not ctx then
    return ""
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
    return ""
  end

  return "[" .. table.concat(parts, " | ") .. "]"
end

return M
