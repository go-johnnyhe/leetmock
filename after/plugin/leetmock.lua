-- after/plugin/leetmock.lua
vim.opt.autoread = true
vim.opt.updatetime = 500
vim.api.nvim_create_autocmd(
	{"FocusGained", "BufEnter", "CursorHold", "CursorHoldI"},
	{ pattern = "*", command = "checktime" }
)