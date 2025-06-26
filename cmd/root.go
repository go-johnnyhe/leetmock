/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "waveland",
	Short: "Instantly share your code editor with anyone, anywhere - no setup, just waveland start",
	Long: `waveland is a real-time collaborative coding tool designed for technical interviews and pair programming.
Share your code instantly with friends, colleagues, or interview partners without any setup or configuration.

Perfect for:
- Mock technical interviews with friends
- Remote pair programming sessions  
- Code reviews and debugging together
- Teaching and mentoring

How it works:
1. Start a session: waveland start main.py
2. Share the generated URL with your partner
3. They join with: waveland join <url>
4. Code together in real-time using your favorite editors

No accounts, no servers to manage, no complex setup - just pure collaborative coding.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		showVersion, _ := cmd.Flags().GetBool("version") 
		if showVersion {
			fmt.Println("waveland version:", Version)
			os.Exit(0)
		}
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().Bool("version", false, "Print the version and exit")

}


