package cmd

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"leetcode/server"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

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
- Share specified files with anyone who joins the session
- Generate a shareable URL for your coding partner

Example:
  (current MVP feature)
  leetmock start main.py              # Share a single file

  (future)
  leetmock start main.py test.py      # Share multiple files  
  leetmock start *.js                 # Share all JavaScript files
  leetmock start .                    # Share current directory

The generated URL can be shared with anyone - they can join using:
  leetmock join <your-session-url>

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

		fmt.Printf("Starting the mock session with %s\n", fileName)

		// Create a context to link with a command line process so that when you stop, we know where to exit
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		// start server in go routine
		http.HandleFunc("/ws", server.StartServer)
		srv := &http.Server{Addr: ":8080"}
		go func() {
			fmt.Println("Websocket server started on :8080")
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				fmt.Printf("Server failed to start: %v\n", err)
				os.Exit(1)
			}
		}()

		// give server a moment to start
		time.Sleep(1 * time.Second)

		fmt.Println("Creating secure tunnel...")
		tunnelURL, err := startCloudflaredTunnel(ctx)
		if err != nil {
			fmt.Printf("Failed to create tunnel: %v\n", err)
			fmt.Println("Server is running locally on localhost:8080")
			return
		}

		fmt.Printf("\n🎉 Session ready!\n")
		fmt.Printf("📡 Session URL: %s\n", tunnelURL)
		fmt.Printf("📁 Sharing: %s\n", fileName)
		fmt.Printf("⚡ Live sync enabled!\n\n")
		fmt.Printf("💡 Your partner should run:\n")
		fmt.Printf("   leetmock join %s\n\n", tunnelURL)
		fmt.Println("🔥 Session active - press Ctrl+C to stop")

		// let the starter user connect as a client too
		go func() {
			time.Sleep(500 * time.Millisecond)
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws",nil)
			if err != nil {
				fmt.Println("Error connecting to websocket: ", err)
				return
			}
			defer conn.Close()

			sendFile(fileName, conn)

			go readFile(conn)
			go monitorFile(conn)

			select{}
		}()

		<-ctx.Done()
		srv.Shutdown(context.Background())
		fmt.Println("")
		fmt.Println("Goodbye!")
		
	},
}

func getCloudflaredBinary() (string, error) {
	binaryName := "cloudflared"
	if runtime.GOOS == "windows" {
		binaryName = "cloudflared.exe"
	}

	binaryPath := binaryName
	if runtime.GOOS != "windows" {
		binaryPath = "./" + binaryName
	}

	if _, err := os.Stat(binaryName); err == nil {
		return binaryPath, nil
	}
	fmt.Println("First time setup: downloading cloudflared (~15MB)...")

	var downloadURL string
	var needsExtraction bool
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64"
	case "linux/arm64":
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64"
	case "darwin/amd64":
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.tgz"
		needsExtraction = true
	case "darwin/arm64":
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-arm64.tgz"
		needsExtraction = true
	case "windows/amd64":
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// download the binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("error downloading cloudflared binary: %v", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %s", resp.Status)
	}

	if needsExtraction {
		if err := extractCloudflaredFromTgz(resp.Body, binaryName); err != nil {
			return "", fmt.Errorf("failed to extract the binary %v", err)
		}
	} else {
		// make binary file
		file, err := os.Create(binaryName)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()
		
		// copy file to binary
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to copy file: %v", err)
		}
	}

	// on unix systems, make binary into executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryName, 0755); err != nil {
			return "", fmt.Errorf("failed to make executable: %v", err)
		}
	}

	fmt.Println("Cloudflared downloaded successfully!")
	return binaryPath, nil
}

func extractCloudflaredFromTgz(reader io.Reader, outputPath string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading from tar header: %v", err)
		}

		if strings.HasSuffix(header.Name, "cloudflared") && header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return fmt.Errorf("error copying binary to output file: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("cloudflared binary not found in the downloaded archive")
}


func startCloudflaredTunnel(ctx context.Context) (string, error){
	binary, err := getCloudflaredBinary()
	if err != nil {
		return "", fmt.Errorf("error getting cloudflared binary: %v", err)
	}


	cmd := exec.CommandContext(ctx, binary, "tunnel", "--url", "localhost:8080")
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create pipe: %v", err)
	// }
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start the command: %v", err)
	}

	// go io.Copy(os.Stderr, stderr)
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		io.Copy(io.MultiWriter(os.Stderr, writer), stderr)
	}()
	scanner := bufio.NewScanner(reader)
	urlRegex := regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.[a-z]+`)

	timeout := time.After(45 * time.Second)
	urlChan := make(chan string, 1)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if match := urlRegex.FindString(line); match != "" {
				urlChan <- match
				return
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "cloudflare scan error: %v\n", err)
		}
	}()

	select {
	case url := <- urlChan:
		return url, nil
	case <-timeout:
		cmd.Process.Kill()
		cmd.Wait()
		return "", fmt.Errorf("timeout waiting for tunnel URL (45s)")
	}

}

func init() {
	rootCmd.AddCommand(startCmd)
}
