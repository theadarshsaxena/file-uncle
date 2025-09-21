/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"text/template"

	"go.uber.org/zap"

	"github.com/theadarshsaxena/file-uncle/internal/config"
	"github.com/theadarshsaxena/file-uncle/internal/src"
)

// var port string
// var username string
// var password string
// var dest string
// var host string

// //go:embed src/static/*
// var staticFiles embed.FS

// //go:embed src/html/upload.html
// var uploadHTML string

// //go:embed src/html/serve.html
// var ServeHTML string

// receiveCmd represents the receive command
// var receiveCmd = &cobra.Command{
// 	Use:   "receive",
// 	Short: "Starts a server to receive files",
// 	Long: `Starts a server to receive files.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		receive()
// 	},
// }

func uploadHandler(uploadDir string) http.HandlerFunc {
	htmlContent := src.UploadHTML
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// tmpl, err := template.ParseFiles("src/html/upload.html")
			tmpl, err := template.New("upload").Parse(htmlContent)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
		} else {
			r.ParseMultipartForm(10 << 20) // limit your max input length!
			file, handler, err := r.FormFile("uploadFile")
			if err != nil {
				fmt.Println("Error retrieving the file")
				fmt.Println(err)
				return
			}
			defer file.Close()

			fmt.Printf("Uploaded File: %s\n", handler.Filename)
			fmt.Printf("File Size: %d\n", handler.Size)
			fmt.Printf("MIME Header: %v\n", handler.Header)

			// if dest != "" {
			// 	if _, err := os.Stat(dest); os.IsNotExist(err) {
			// 		fmt.Println("Destination folder does not exist")
			// 		return
			// 	}
			// 	if dest[len(dest)-1:] == "/" {
			// 		dest = dest[:len(dest)-1]
			// 	}
			// } else{
			// 	dest = "./uploads"
			// }
			dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
			if err != nil {
				fmt.Println("Error creating file")
				fmt.Println(err)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				fmt.Println("Error copying file")
				fmt.Println(err)
				return
			}
			fmt.Fprintf(w, "Successfully Uploaded File\n")
		}
	}
}

// func getLocalIP() (string, error) {
// 	addrs, err := net.InterfaceAddrs()
// 	if err != nil {
// 		return "", err
// 	}

// 	address := ""

// 	for _, addr := range addrs {
// 		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
// 			if ipnet.IP.To4() != nil {
// 				fmt.Println(ipnet.IP.String())
// 				address = ipnet.IP.String()
// 				// return ipnet.IP.String(), nil
// 			}
// 		}
// 	}
// 	if address != "" {
// 		return address, nil
// 	}
// 	return "", fmt.Errorf("cannot find local IP address")
// }

func RunReceive(logger *zap.Logger) error {
	if config.Shared.Host == "" {
		config.Shared.Host = "localhost"
	}
	// Get the current user's home directory
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return err
	}
	uploadDir := filepath.Join(usr.HomeDir, "uploads")

	// Create the uploads directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.Mkdir(uploadDir, 0755)
		if err != nil {
			fmt.Println("Error creating uploads directory:", err)
			return err
		}
	}

	// Print the destination folder
	fmt.Printf("Destination folder: %s\n", uploadDir)
    // Serve static files
	staticFs, err := fs.Sub(src.StaticFiles, "src/static")
	if err != nil {
		fmt.Println("Error serving static files:", err)
		return err
	}
    fs := http.FileServer(http.FS(staticFs))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

	// http.Handle("/static/", http.FileServer(http.FS(staticFiles)))
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "src/html/upload.html")
	// })

	if config.Shared.Username != "" && config.Shared.Password != "" {
		http.Handle("/", basicAuth(uploadHandler(uploadDir)))
	} else {
		http.HandleFunc("/", uploadHandler(uploadDir))
	}
	if config.Shared.Username == "" && config.Shared.Password == "" {
		fmt.Println("Authentication disabled (password and username not provided)")
	} else {
		fmt.Println("Authentication enabled with username: " + config.Shared.Username + " and password: " + config.Shared.Password)
	}

	fmt.Println("\nServer started on: http://" + config.Shared.Host + ":" + config.Shared.Port)
	
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if withNgrok {
		address := fmt.Sprintf("http://localhost:%s", config.Shared.Port)
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
	http.ListenAndServe(config.Shared.Host + ":" + config.Shared.Port, nil)
	return nil
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != config.Shared.Username || pass != config.Shared.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	// rootCmd.AddCommand(receiveCmd)
	
}