/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "leetmock",
	Short: "Instantly share your code editor with anyone, anywhere - no setup, just leetmock start",
	Long: `leetmock is a real-time collaborative coding tool designed for technical interviews and pair programming.
Share your code instantly with friends, colleagues, or interview partners without any setup or configuration.

Perfect for:
- Mock technical interviews with friends
- Remote pair programming sessions  
- Code reviews and debugging together
- Teaching and mentoring

How it works:
1. Start a session: leetmock start main.py
2. Share the generated URL with your partner
3. They join with: leetmock join <url>
4. Code together in real-time using your favorite editors

No accounts, no servers to manage, no complex setup - just pure collaborative coding.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.leetcode.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


