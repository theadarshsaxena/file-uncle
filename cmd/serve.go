/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"unicode/utf8"

	"github.com/spf13/cobra"
	ngrok "golang.ngrok.com/ngrok/v2"
)

 var (
	directory string
	generatedKey []byte
 )

// You may need to define trafficPolicy if not already present
// var trafficPolicy ngrok.TrafficPolicy

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
	Path string
	PathEncrypted string
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

// EncryptDeterministic provides consistent output for same input
func EncryptDeterministic(plaintext string, key []byte) (string, error) {
    if len(key) != 32 {
        return "", fmt.Errorf("key must be exactly 32 bytes for AES-256")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }

    // Generate deterministic IV from plaintext hash
    hash := sha256.Sum256([]byte(plaintext + string(key[:16])))
    iv := hash[:16] // Use first 16 bytes as IV

    // Pad plaintext to block size
    paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

    ciphertext := make([]byte, len(paddedPlaintext))
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext, paddedPlaintext)

    // Prepend IV to ciphertext
    result := append(iv, ciphertext...)
    return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptDeterministic decrypts deterministically encrypted data
func DecryptDeterministic(encrypted string, key []byte) (string, error) {
    if len(key) != 32 {
        return "", fmt.Errorf("key must be exactly 32 bytes for AES-256")
    }

    data, err := base64.StdEncoding.DecodeString(encrypted)
    if err != nil {
        return "", fmt.Errorf("failed to decode base64: %w", err)
    }

    if len(data) < aes.BlockSize {
        return "", fmt.Errorf("encrypted data too short")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }

    // Extract IV and ciphertext
    iv := data[:aes.BlockSize]
    ciphertext := data[aes.BlockSize:]

    if len(ciphertext)%aes.BlockSize != 0 {
        return "", fmt.Errorf("ciphertext length not multiple of block size")
    }

    plaintext := make([]byte, len(ciphertext))
    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(plaintext, ciphertext)

    // Remove padding
    unpaddedPlaintext, err := pkcs7Unpad(plaintext)
    if err != nil {
        return "", fmt.Errorf("failed to remove padding: %w", err)
    }

    return string(unpaddedPlaintext), nil
}

// PKCS7 padding functions
func pkcs7Pad(data []byte, blockSize int) []byte {
    padding := blockSize - (len(data) % blockSize)
    padText := make([]byte, padding)
    for i := range padText {
        padText[i] = byte(padding)
    }
    return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
    if len(data) == 0 {
        return nil, fmt.Errorf("empty data")
    }
    
    padding := int(data[len(data)-1])
    if padding > len(data) || padding == 0 {
        return nil, fmt.Errorf("invalid padding")
    }
    
    for i := len(data) - padding; i < len(data); i++ {
        if data[i] != byte(padding) {
            return nil, fmt.Errorf("invalid padding")
        }
    }
    
    return data[:len(data)-padding], nil
}

// GenerateKey creates a cryptographically secure random key
func GenerateKey(size int) ([]byte, error) {
    key := make([]byte, size)
    _, err := rand.Read(key)
    return key, err
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	folderPath := directory

	var files []FileInfo

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			encryptedPath, err := EncryptDeterministic(path, generatedKey)
			if err != nil {
				return err
			}
			displayName := truncateFileName(info.Name(), 30)
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			files = append(files, FileInfo{
				Name: info.Name(),
				DisplayName: displayName,
				FileSize: FileSize(info.Size()),
				LineNumber: len(files) + 1,
				Path: absPath,
				PathEncrypted: encryptedPath,
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
	encryptedFileName := r.URL.Query().Get("file")
	encryptedFileName = strings.ReplaceAll(encryptedFileName, " ", "+")
	fmt.Println("Encrypted file parameter received:", encryptedFileName)
	fileName, err := DecryptDeterministic(encryptedFileName, generatedKey)
	if err != nil {
		fmt.Println("Decryption error:", err)
		http.Error(w, "Invalid file parameter", http.StatusBadRequest)
		return
	}
	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Println("Someone tried downloading file: ", fileName)
	http.ServeFile(w, r, fileName)
}

var withNgrok bool

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

func serveFile() {
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download/", downloadFile)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/src/static"))))

	fmt.Println("Server starting at http://localhost:8080")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if withNgrok {
		address := fmt.Sprintf("http://%s:%s", host, port)
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

	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil)
}

func init() {
    var err error
    generatedKey, err = GenerateKey(32)
    if err != nil {
        panic("Failed to generate key: " + err.Error())
    }
	// receiveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port number for the server")
	// receiveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host address or Local IP to bind the server to (default is localhost)")
	receiveCmd.Flags().StringVarP(&directory, "directory", "y", "./", "Directory to serve files from")
	serveCmd.Flags().BoolVar(&withNgrok, "with-ngrok", false, "Start an ngrok tunnel for public access")
	rootCmd.AddCommand(serveCmd)
}
