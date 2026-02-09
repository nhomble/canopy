local M = {}

--- Open a floating window with the given lines.
--- @param lines string[] Content lines
--- @param title string Window title
--- @param opts table|nil Extra options { on_cr = function(line_nr) }
--- @return number buf, number win
local function open_float(lines, title, opts)
  opts = opts or {}

  -- Calculate dimensions
  local max_width = #title + 4
  for _, line in ipairs(lines) do
    max_width = math.max(max_width, #line + 2)
  end
  local width = math.min(max_width, math.floor(vim.o.columns * 0.8))
  local height = math.min(#lines, math.floor(vim.o.lines * 0.6))

  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)

  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)

  vim.bo[buf].modifiable = false
  vim.bo[buf].bufhidden = "wipe"
  vim.bo[buf].buftype = "nofile"

  local win = vim.api.nvim_open_win(buf, true, {
    relative = "editor",
    row = row,
    col = col,
    width = width,
    height = height,
    style = "minimal",
    border = "rounded",
    title = " " .. title .. " ",
    title_pos = "center",
  })

  -- Close keymaps
  local close = function()
    if vim.api.nvim_win_is_valid(win) then
      vim.api.nvim_win_close(win, true)
    end
  end
  vim.keymap.set("n", "q", close, { buffer = buf, nowait = true })
  vim.keymap.set("n", "<Esc>", close, { buffer = buf, nowait = true })

  -- Optional CR handler
  if opts.on_cr then
    vim.keymap.set("n", "<CR>", function()
      local line_nr = vim.api.nvim_win_get_cursor(win)[1]
      close()
      opts.on_cr(line_nr)
    end, { buffer = buf, nowait = true })
  end

  return buf, win
end

--- Show the full architectural context in a floating window.
--- @param ctx table ContextResponse from server
function M.show_context(ctx)
  local lines = {}

  if ctx.component then
    table.insert(lines, "Component: " .. (ctx.component.name or ctx.component.id))
  end
  if ctx.layer and ctx.layer ~= "" then
    table.insert(lines, "Layer:     " .. ctx.layer)
  end
  if ctx.archetype then
    local arch_str = ctx.archetype.category or ""
    if ctx.archetype.technology and ctx.archetype.technology ~= "" then
      arch_str = arch_str .. " (" .. ctx.archetype.technology .. ")"
    end
    table.insert(lines, "Archetype: " .. arch_str)
  end

  if ctx.flows and #ctx.flows > 0 then
    table.insert(lines, "")
    table.insert(lines, "Flows:")
    for _, flow in ipairs(ctx.flows) do
      table.insert(lines, "  * " .. flow.name)
    end
  end

  if #lines == 0 then
    table.insert(lines, "(no architectural context)")
  end

  open_float(lines, "Architectural Context")
end

--- Show a picker for selecting a flow.
--- @param flows table[] Array of { id, name }
--- @param on_select function(flow) Called with the selected flow
function M.pick_flow(flows, on_select)
  local lines = {}
  for i, flow in ipairs(flows) do
    table.insert(lines, string.format(" %d. %s", i, flow.name))
  end

  open_float(lines, "Select Flow", {
    on_cr = function(line_nr)
      local flow = flows[line_nr]
      if flow then
        on_select(flow)
      end
    end,
  })
end

--- Show the steps of a flow with file paths for jumping.
--- @param flow table Flow object with steps array
--- @param archetype_files table<string, string> archetype_id → file_path
--- @param project_root string|nil Absolute path to project root
function M.show_flow_steps(flow, archetype_files, project_root)
  local lines = {}
  local step_files = {}

  for i, step_id in ipairs(flow.steps) do
    local file = archetype_files[step_id]
    local display = string.format(" %d. %s", i, step_id)
    if file then
      display = display .. "  -> " .. file
    end
    table.insert(lines, display)
    step_files[i] = file
  end

  open_float(lines, flow.name, {
    on_cr = function(line_nr)
      local file = step_files[line_nr]
      if not file then
        vim.notify("canopy: no file for this step", vim.log.levels.WARN)
        return
      end
      -- Resolve to absolute path if we have a project root
      if project_root then
        file = project_root .. "/" .. file
      end
      vim.cmd("edit " .. vim.fn.fnameescape(file))
    end,
  })
end

return M
