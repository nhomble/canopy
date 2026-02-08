local client = require("arch-index.client")

local M = {}

-- Per-buffer context cache: bufnr → { path = string, context = table }
local buf_cache = {}

-- Session-level archetype file map: archetype_id → file_path
local archetype_files = nil

--- Detect the project root by walking up from the buffer's directory
--- looking for a .arch/ directory.
--- @param bufnr number
--- @return string|nil root Absolute path to project root
local function find_project_root(bufnr)
  local bufpath = vim.api.nvim_buf_get_name(bufnr)
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

--- Compute the relative path the server expects (forward slashes, no leading ./ or /).
--- @param bufnr number
--- @return string|nil relpath
function M.relative_path(bufnr)
  local bufpath = vim.api.nvim_buf_get_name(bufnr)
  if bufpath == "" then
    return nil
  end

  local root = find_project_root(bufnr)
  if not root then
    return nil
  end

  -- Strip root prefix and normalize
  local rel = bufpath:sub(#root + 2) -- +2 to skip the trailing /
  rel = rel:gsub("\\", "/")
  rel = rel:gsub("^%./", "")
  rel = rel:gsub("^/", "")
  return rel
end

--- Get context for a buffer, using cache if available.
--- @param bufnr number
--- @param base_url string
--- @return table|nil context
function M.get(bufnr, base_url)
  local rel = M.relative_path(bufnr)
  if not rel then
    return nil
  end

  -- Check cache: valid if same file path
  local cached = buf_cache[bufnr]
  if cached and cached.path == rel then
    return cached.context
  end

  -- Fetch from server
  local url = base_url .. "/context?file=" .. vim.uri_encode(rel, "rfc2396")
  local data, err = client.get(url)
  if not data then
    return nil
  end

  buf_cache[bufnr] = { path = rel, context = data }
  return data
end

--- Invalidate cache for a buffer (called on BufEnter when file changes).
--- @param bufnr number
function M.invalidate(bufnr)
  local rel = M.relative_path(bufnr)
  local cached = buf_cache[bufnr]
  if cached and cached.path ~= rel then
    buf_cache[bufnr] = nil
  end
end

--- Get the archetype_id → file_path map, fetching /graph once per session.
--- @param base_url string
--- @return table<string, string> map
function M.get_archetype_files(base_url)
  if archetype_files then
    return archetype_files
  end

  local data, err = client.get(base_url .. "/graph")
  if not data or not data.components then
    return {}
  end

  archetype_files = {}
  for _, comp in ipairs(data.components) do
    if comp.archetypes then
      for _, arch in ipairs(comp.archetypes) do
        archetype_files[arch.id] = arch.file
      end
    end
  end

  return archetype_files
end

return M
