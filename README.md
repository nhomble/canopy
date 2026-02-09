# arch-index

CLI tool that uses LLMs to analyze codebases and generate architectural context.

## Usage

```bash
arch-index init
arch-index prepare-analysis . | claude --print | arch-index import --force
arch-index serve
```
