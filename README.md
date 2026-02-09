# canopy

see the forest from the trees

## About

CLI tool that uses LLMs to analyze codebases and generate architectural context. It is meant to be
viewed side by side during development.

## Usage

```bash
canopy init
canopy prepare-analysis . | claude --print | canopy import --force
canopy serve
```

## Install

```bash
go install github.com/nhomble/canopy/cmd/canopy@latest
```

### Neovim (lazy.nvim)

```lua
{
  "nhomble/canopy",
  config = function()
    require("canopy").setup({
      binary = vim.fn.exepath("canopy"), -- or absolute path
    })
  end,
}
```

Commands: `:CanopyStart`, `:CanopyStop`, `:CanopyRestart`, `:CanopyContext`, `:CanopyFlow`, `:CanopyOpen`

## Demo

![](./docs/demo.png)
