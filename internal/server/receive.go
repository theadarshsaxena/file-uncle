/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"go.uber.org/zap"

	"github.com/theadarshsaxena/file-uncle/internal/config"
	"github.com/theadarshsaxena/file-uncle/internal/logging"
	"github.com/theadarshsaxena/file-uncle/internal/ngrok"
	"github.com/theadarshsaxena/file-uncle/internal/src"
)

var logger *zap.Logger

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
			files := r.MultipartForm.File["uploadFile"]
			
			if len(files) == 0 {
				fmt.Println("Error: no files uploaded")
				http.Error(w, "No files uploaded", http.StatusBadRequest)
				return
			}

			totalSize := int64(0)
			successCount := 0
			failureCount := 0

			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					fmt.Printf("Error opening file %s: %v\n", fileHeader.Filename, err)
					failureCount++
					continue
				}
				defer file.Close()

				// Extract the file path (handles both single files and folder structure)
				filePath := fileHeader.Filename
				
				// Create destination path, preserving folder structure
				dstPath := filepath.Join(uploadDir, filePath)
				dstDir := filepath.Dir(dstPath)

				// Create directories if they don't exist
				if err := os.MkdirAll(dstDir, 0755); err != nil {
					fmt.Printf("Error creating directory %s: %v\n", dstDir, err)
					failureCount++
					continue
				}

				// Create the file
				dst, err := os.Create(dstPath)
				if err != nil {
					fmt.Printf("Error creating file %s: %v\n", dstPath, err)
					failureCount++
					continue
				}
				defer dst.Close()

				// Copy file contents
				if _, err := io.Copy(dst, file); err != nil {
					fmt.Printf("Error copying file %s: %v\n", filePath, err)
					failureCount++
					continue
				}

				logging.LogReceive(filePath, fileHeader.Size, r.RemoteAddr)
				totalSize += fileHeader.Size
				successCount++
			}

			if failureCount > 0 {
				fmt.Fprintf(w, "Uploaded %d files successfully with %d failures\n", successCount, failureCount)
			} else {
				fmt.Fprintf(w, "Successfully Uploaded %d files\n", successCount)
			}
		}
	}
}

func getPhysicalInterfaceIP() (string, error) {
    interfaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    
    // Skip common virtual interface patterns
    skipPatterns := []string{"docker", "veth", "br-", "virbr", "vmnet", "vbox", "lo"}
	address := ""
    
    for _, iface := range interfaces {
        // Skip virtual interfaces
        isVirtual := false
        for _, pattern := range skipPatterns {
            if strings.Contains(strings.ToLower(iface.Name), pattern) {
                isVirtual = true
                break
            }
        }
        if isVirtual {
            continue
        }
        
        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }

		
        
        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                if ipv4 := ipnet.IP.To4(); ipv4 != nil {
                    address = ipv4.String()
					// fmt.Println("Using physical interface:", iface.Name, "with IP:", address)  // TODO: log this with debug level later
					return address, nil
                }
            }
        }
    }
    
    return "", fmt.Errorf("no physical interface IP found")
}

func RunReceive(logger *zap.Logger) error {
	fmt.Println("🚀 File-Uncle Server Starting...")
	fmt.Println()
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

	fmt.Printf("📁 Upload Directory: %s\n", uploadDir)
	staticFS, err := fs.Sub(src.StaticFiles, "static")
	if err != nil {
		panic(err)
	}	
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staticHandler.ServeHTTP(w, r)
	}))

	if config.Shared.Username != "" && config.Shared.Password != "" {
		http.Handle("/", basicAuth(uploadHandler(uploadDir)))
	} else {
		http.HandleFunc("/", uploadHandler(uploadDir))
	}
	if config.Shared.Username == "" && config.Shared.Password == "" {
		fmt.Println("🔓 Authentication: Disabled (no credentials provided)")
	} else {
		fmt.Println("🔓 Authentication enabled with username: " + config.Shared.Username + " and password: " + config.Shared.Password)
	}

	localIP, err := getPhysicalInterfaceIP()
	if err != nil {
		fmt.Println("Error getting physical interface IP, falling back to localhost:", err)
		localIP = "localhost"
	}
	config.Shared.Host = "0.0.0.0"
	fmt.Println("🌐 Starting server on http://" + config.Shared.Host + ":" + config.Shared.Port)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\n🏠 Local Network Access: \033[33mhttp://%s:%s\033[0m\n", localIP, config.Shared.Port)

	// fmt.Println("\nServer started on: http://" + config.Shared.Host + ":" + config.Shared.Port)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urlChan := make(chan string, 1)
	errChan := make(chan error, 1)

	if config.Shared.WithNgrok {
		// check if the environment variable is set
		address := fmt.Sprintf("http://%s:%s", localIP, config.Shared.Port)
		go ngrok.RunNgrok(ctx, address, urlChan, errChan)
		select {
		case url := <-urlChan:
			fmt.Printf("🌍 Public Access (ngrok): \033[33m%s\033[0m\n", url)
		case err := <-errChan:
			fmt.Printf("\033[31mngrok error: %s\033[0m\n", err)
			fmt.Printf("Note: Any device on your network can still access reach here using the local network address above\n")
		}
	}
	
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	fmt.Println("To send files, open the above URL in your browser or use command: curl -F 'uploadFile=@/path/to/your/file' <URL>")
	
	go func() {
		<-sigs
		fmt.Println("Stopped local http server and also ngrok tunnel stopped (if enabled)")
		cancel()
		os.Exit(0)
	}()
	http.ListenAndServe(config.Shared.Host+":"+config.Shared.Port, nil)
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
	logger = logging.GetLogger(config.Shared.LogLevel)
}
