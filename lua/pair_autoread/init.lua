local M = {}

function M.setup(opts)
    opts = opts or {}
    vim.o.autoread = true
    vim.o.updatetime = opts.updatetime or 300

    local aug = vim.api.nvim_create_augroup("PairAutoRead", { clear = true })
    local function safe_check()
        if not vim.bo.modified then pcall(vim.cmd, "checktime") end
    end

    vim.api.nvim_create_autocmd(
        { "CursorHold", "CursorHoldI", "FocusGained", "BufEnter",
          "ModeChanged", "TextChanged", "TextChangedI"},
        { group = aug, callback = safe_check }
    )

    vim.api.nvim_create_autocmd("FileChangedShellPost", {
        group = aug,
        callback = function()
            vim.notify("Buffer reloaded from disk", vim.log.levels.INFO, { title = "PairAutoRead" })
        end,
    })
end

return M