/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [file...]",
	Short: "Start a collaborative coding session and share files",
	Long: `Start a new collaborative coding session with instant file sharing.

This command will:
- Launch a WebSocket server for real-time collaboration
- Create a secure tunnel using ngrok (no setup required)
- Share specified files with anyone who joins the session
- Generate a shareable URL for your coding partner

Examples:
  leetmock start main.py              # Share a single file
  leetmock start main.py test.py      # Share multiple files  
  leetmock start *.js                 # Share all JavaScript files
  leetmock start .                    # Share current directory

The generated URL can be shared with anyone - they can join using:
  leetmock join <your-session-url>

Perfect for mock interviews, pair programming, and collaborative debugging.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
