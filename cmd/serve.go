/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
	"unicode/utf8"

	"github.com/spf13/cobra"
	ngrok "golang.ngrok.com/ngrok/v2"
)

// You may need to define trafficPolicy if not already present
// var trafficPolicy ngrok.TrafficPolicy

func runNgrok(ctx context.Context, address string) error {
	ngrokAuthToken := os.Getenv("NGROK_AUTHTOKEN")
	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(ngrokAuthToken))
	if err != nil {
		return err
	}

	ln, err := agent.Forward(ctx,
		ngrok.WithUpstream(address),
		ngrok.WithURL(os.Getenv("NGROK_RESERVED_DOMAIN")),
		// ngrok.WithTrafficPolicy(trafficPolicy),
	)

	if err != nil {
		fmt.Println("Error", err)
		return err
	}

	fmt.Println("Endpoint online: forwarding from", ln.URL(), "to", address)

	// Explicitly stop forwarding; otherwise it runs indefinitely
	<-ln.Done()
	return nil
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		serveFile()
	},
}

type FileInfo struct {
	LineNumber int
	Name string
	DisplayName string
	FileSize string
}

func truncateFileName(name string, length int) string {
	if utf8.RuneCountInString(name) > length {
		return name[:length] + "..."
	}
	return name
}

func FileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	folderPath := "./" // folder to serve files from

	var files []FileInfo

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			displayName := truncateFileName(info.Name(), 30)
			files = append(files, FileInfo{
				Name: info.Name(),
				DisplayName: displayName,
				FileSize: FileSize(info.Size()),
				LineNumber: len(files) + 1,
			})
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("serve").Parse(ServeHTML)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
  
	tmpl.Execute(w, files)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	filePath := fmt.Sprintf("./files/%s", fileName)

	http.ServeFile(w, r, filePath)
}

var withNgrok bool

func serveFile() {
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download", downloadFile)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/src/static"))))

	fmt.Println("Server started at http://localhost:8080")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if withNgrok {
		address := fmt.Sprintf("http://localhost:%s", port)
		go func() {
			err := runNgrok(ctx, address)
			if err != nil {
				fmt.Println("ngrok error:", err)
			}
		}()
	}

	go func() {
		<-sigs
		fmt.Println("Stopped local http server and also ngrok tunnel stopped (if enabled)")
		cancel()
		os.Exit(0)
	}()

	http.ListenAndServe("localhost:8080", nil)
}

func init() {
	// receiveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port number for the server")
	// receiveCmd.Flags().StringVarP(&host, "host", "H", "", "Host address or Local IP to bind the server to (default is localhost)")
	serveCmd.Flags().BoolVar(&withNgrok, "with-ngrok", false, "Start an ngrok tunnel for public access")
	rootCmd.AddCommand(serveCmd)
}
