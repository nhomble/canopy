-- Guard against double-load
if vim.g.loaded_arch_index then
  return
end
vim.g.loaded_arch_index = true

-- Defer setup so users can call require('arch-index').setup(opts) first.
-- If they haven't called setup by VimEnter, set up with defaults.
vim.api.nvim_create_autocmd("VimEnter", {
  group = vim.api.nvim_create_augroup("arch-index-bootstrap", { clear = true }),
  once = true,
  callback = function()
    require("arch-index").setup()
  end,
})
