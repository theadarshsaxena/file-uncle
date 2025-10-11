/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package server

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"unicode/utf8"

	"github.com/theadarshsaxena/file-uncle/internal/config"
	"github.com/theadarshsaxena/file-uncle/internal/logging"
	"github.com/theadarshsaxena/file-uncle/internal/ngrok"
	"github.com/theadarshsaxena/file-uncle/internal/src"
	"go.uber.org/zap"
)

var (
	generatedKey []byte
)

type FileInfo struct {
	LineNumber    int
	Name          string
	DisplayName   string
	FileSize      string
	Path          string
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
	folderPath := config.Shared.Directory

	var files []FileInfo

	// TODO: Add pagination for large number of files, or limit to certain number of files
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
				Name:          info.Name(),
				DisplayName:   displayName,
				FileSize:      FileSize(info.Size()),
				LineNumber:    len(files) + 1,
				Path:          absPath,
				PathEncrypted: encryptedPath,
			})
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}
	serveHtml := src.ServeHTML
	tmpl, err := template.New("serve").Parse(serveHtml)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, files)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	encryptedFileName := r.URL.Query().Get("file")
	encryptedFileName = strings.ReplaceAll(encryptedFileName, " ", "+")
	fileName, err := DecryptDeterministic(encryptedFileName, generatedKey)
	if err != nil {
		fmt.Println("Decryption error:", err)
		http.Error(w, "Invalid file parameter", http.StatusBadRequest)
		return
	}
	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	fileSize := fileInfo.Size()
	logging.LogServe(fileName, fileSize, r.RemoteAddr)
	http.ServeFile(w, r, fileName)
}

func RunServe(logger *zap.Logger) error {
	fmt.Println("🚀 File-Uncle Server Starting...")
	
	fmt.Printf("\n╭─────────────────────────────────────────────────────────────────────────────╮\n")
	staticFS, err := fs.Sub(src.StaticFiles, "static")
	if err != nil {
		panic(err)
	}	
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staticHandler.ServeHTTP(w, r)
	}))

	http.HandleFunc("/download/", downloadFile)
	http.HandleFunc("/", listFiles)

	localIP, err := getPhysicalInterfaceIP()
	if err != nil {
		fmt.Println("│ • Error getting physical interface IP, falling back to localhost:", err)
		localIP = "localhost"
	}
	config.Shared.Host = "0.0.0.0"

	fmt.Printf("│ • Serving files from Directory: %-44s│\n", config.Shared.Directory)
	fmt.Printf("│ • Authentication not supported in serve cmd yet %28s│\n", "")
	
	fmt.Println("╰─────────────────────────────────────────────────────────────────────────────╯")
	fmt.Printf("🌐 Starting server on http://%s:%s\n", config.Shared.Host, config.Shared.Port)
	fmt.Printf("\n╭─────────────────────────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│ Access URLs                                                                 │\n")
	fmt.Printf("│ • Local Network Access: \033[33mhttp://%-45s\033[0m│\n", fmt.Sprintf("%s:%s",localIP, config.Shared.Port))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	

	if config.Shared.WithNgrok {
		urlChan := make(chan string, 1)
		errChan := make(chan error, 1)
		address := fmt.Sprintf("http://%s:%s", localIP, config.Shared.Port)
		go ngrok.RunNgrok(ctx, address, urlChan, errChan)
		select {
		case url := <-urlChan:
			fmt.Printf("│ • Public URL available: %20s│\n", url)
		case err := <-errChan:
			fmt.Printf("│\033[31m • ngrok error: %s\033[0m│\n", err)
			fmt.Printf("│ • Note: Any device on your network can still access reach here using the local network address above│\n")
		}
	} else {
		fmt.Printf("│ • Public access disabled, use --with-ngrok in command to start tunnel %6s│\n", "")
	}

	fmt.Println("╰─────────────────────────────────────────────────────────────────────────────╯")

	go func() {
		<-sigs
		fmt.Println("Stopped local http server and also ngrok tunnel stopped (if enabled)")
		cancel()
		os.Exit(0)
	}()
	http.ListenAndServe(fmt.Sprintf("%s:%s", config.Shared.Host, config.Shared.Port), nil)
	return nil
}

func init() {
	var err error
	generatedKey, err = GenerateKey(32)
	if err != nil {
		panic("Failed to generate key: " + err.Error())
	}
}
