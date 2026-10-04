
vim.lsp.config("gopls", {
  settings = {
    gopls = {
      buildFlags = {"-tags=sqlite postgres mysql redis"}
    }
  }
})
