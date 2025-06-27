package cmd

import (
	"context"
	"fmt"
	"github.com/go-johnnyhe/waveland/server"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"github.com/go-johnnyhe/waveland/internal/client"
	"github.com/go-johnnyhe/waveland/internal/tunnel"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)


// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <file>",
	Short: "Start a collaborative coding session and share files",
	Long: `Start a new collaborative coding session with instant file sharing.

This command will:
- Launch a WebSocket server for real-time collaboration
- Create a secure tunnel using Cloudflared (no setup required)
- Share the current directory or specified files with anyone who joins the session
- Generate a shareable URL for your coding partner

Example:
  waveland start main.py              # Share a single file
  waveland start .                    # Share current directory

The generated URL can be shared with anyone - they can join using:
  waveland join <your-session-url>

Perfect for mock interviews, pair programming, and collaborative debugging.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Error: this takes exactly one file")
			cmd.Usage()
			return
		}

		fileName := args[0]

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if f, err := os.Create(fileName); err != nil {
				fmt.Printf("failed to create %s: %v\n", fileName, err)
				return
			} else {
				f.Close()
				fmt.Printf("Created %s (empty file)\n", fileName)
			}
		} else if err != nil {
			fmt.Printf("error checking %s: %v\n", fileName, err)
			return
		}

		// fmt.Printf("Starting the mock session with %s\n", fileName)

		// Create a context to link with a command line process so that when you stop, we know where to exit
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		// start server in go routine
		http.HandleFunc("/ws", server.StartServer)
		srv := &http.Server{Addr: ":8080"}
		go func() {
			// fmt.Println("Websocket server started on :8080")
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				if strings.Contains(err.Error(), "address already in use") {
					fmt.Println("Port 8080 is already in use, please close other applications using this port.")
					fmt.Println("\nTo find what's using port 8080:")
					fmt.Println("  Linux/Mac: lsof -i :8080")
					fmt.Println("  Windows: netstat -ano | findstr :8080")
				}
				fmt.Printf("Server failed to start: %v\n", err)
				os.Exit(1)
			}
		}()

		// give server a moment to start
		time.Sleep(1 * time.Second)

		// fmt.Println("Connecting...")
		tunnelURL, err := tunnel.StartCloudflaredTunnel(ctx)
		if err != nil {
			fmt.Printf("Failed to create tunnel: %v\n", err)
			fmt.Println("Server is running locally on localhost:8080")
			return
		}

		fmt.Printf("\n✅ Wavelanding %s\n", fileName)
		fmt.Println("")
		fmt.Printf("Share this command with your partner:\n")

		// Bold the command for better visibility
		if os.Getenv("TERM") != "dumb" && os.Getenv("NO_COLOR") == "" {
			fmt.Printf("\n  \033[1mwaveland join %s\033[0m\n", tunnelURL)
		} else {
			fmt.Printf("\n  waveland join %s\n", tunnelURL)
		}

		// let the starter user connect as a client too
		go func(ctx context.Context) {
			time.Sleep(500 * time.Millisecond)
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws",nil)
			if err != nil {
				fmt.Println("Error connecting to websocket: ", err)
				return
			}
			defer conn.Close()

			c := client.NewClient(conn)
			c.SendFile(fileName)
			c.Start(ctx)
			<-ctx.Done()
		}(ctx)

		<-ctx.Done()
		srv.Shutdown(context.Background())
		time.Sleep(100 * time.Millisecond)
		fmt.Println("")
		fmt.Println("Goodbye!")
		
	},
}



func init() {
	rootCmd.AddCommand(startCmd)
}
