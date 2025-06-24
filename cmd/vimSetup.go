/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed extras/autoread.vim
var pluginBody []byte

const luaSnippet = `-- ~/.config/nvim/after/plugin/leetmock.lua
vim.opt.autoread = true
vim.opt.updatetime = 500
local group = vim.api.nvim_create_augroup("leetmock_autoread", { clear = true })

vim.api.nvim_create_autocmd(
  { "FocusGained", "BufEnter", "CursorHold", "CursorHoldI", "TermEnter" },
  {
    group = group,
    pattern = "*",
    callback = function()
      -- pcall avoids 'checktime' errors in special buffers
      pcall(vim.cmd, "checktime")
    end,
    desc = "Reload buffer if the file changed on disk",
  }
)`

var vimSetupCmd = &cobra.Command{
	Use:   "vimSetup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		failures := []string{}
		var ok bool

		if data, err := nvimDataDir(); err == nil {
			dst := filepath.Join(data,
			"site", "pack", "leetmock", "start",
			"autoread", "plugin", "autoread.vim")
			if err := copyFile(dst, pluginBody); err == nil {
				fmt.Println("✅ Neovim config is done (data path)")
				ok = true
			} else {
				failures = append(failures, "Neovim-data: " + err.Error())
			}
		}

		// nvim config/plugin fallback
		if cfg, err := nvimConfigDir(); err == nil {
			if err := installNvimScriptAfterPlugin(cfg); err == nil {
				fmt.Println("✅ Neovim config is done (after/plugin path), restart your nvim")
				ok = true
			} else {
				failures = append(failures, "Neovim-cfg: " + err.Error())
			}
		}

		configDst := filepath.Join(vimSiteDir(), 
		"pack", "leetmock", "start",
		"autoread", "plugin", "autoread.vim")
		if err := copyFile(configDst, pluginBody); err == nil {
			fmt.Println("✅ Vim config is done, restart your vim")
			ok = true
		} else {
			failures = append(failures, "Vim: " + err.Error())
		}
		
			
		
		if !ok {
			return fmt.Errorf("%s", strings.Join(failures, "; "))
		}

		return nil
	},
}

func nvimDataDir() (string, error) {
	cmd := exec.Command("nvim", "--headless", "--clean", "+lua print(vim.fn.stdpath('data'))", "+q")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func nvimConfigDir() (string, error) {
	cmd := exec.Command("nvim", "--headless", "--clean", "+lua print(vim.fn.stdpath('config'))", "+q")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func installNvimScriptAfterPlugin(cfg string) error {
	dst := filepath.Join(cfg, "after", "plugin", "leetmock.lua")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(luaSnippet), 0o644)
}

func vimSiteDir() (string) {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vim")
}


func copyFile(dest string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dest, body, 0o644)
}


func init() {
	rootCmd.AddCommand(vimSetupCmd)
}
